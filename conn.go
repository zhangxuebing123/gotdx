package gotdx

import (
	"context"
	. "gotdx/imsg"
	"gotdx/logger"
	"net"
	"sync"
	"time"
)

const (
	MessageHeaderBytes = 0x10
	MessageMaxBytes    = 1 << 15
)

const (
	RECONNECT_INTERVAL = 3 // 重连时间
)

type options struct {
	codec      Codec
	onConnect  onConnectFunc
	onClose    onCloseFunc
	onError    onErrorFunc
	workerSize int  // numbers of worker go-routines
	bufferSize int  // size of buffered channel
	reconnect  bool // for ClientConn use only
}

type Option func(*options)

func ReconnectOption() Option {
	return func(o *options) {
		o.reconnect = true
	}
}

func CustomCodecOption(codec Codec) Option {
	return func(o *options) {
		o.codec = codec
	}
}

func WorkerSizeOption(workerSz int) Option {
	return func(o *options) {
		o.workerSize = workerSz
	}
}

func BufferSizeOption(indicator int) Option {
	return func(o *options) {
		o.bufferSize = indicator
	}
}

func OnConnectOption(cb func(WriteCloser) bool) Option {
	return func(o *options) {
		o.onConnect = cb
	}
}

func OnCloseOption(cb func(WriteCloser)) Option {
	return func(o *options) {
		o.onClose = cb
	}
}

func OnErrorOption(cb func(WriteCloser)) Option {
	return func(o *options) {
		o.onError = cb
	}
}

func cancelTimer(timing *TimingWheel, timerID int64) {
	if timing != nil {
		timing.CancelTimer(timerID)
	}
}

type WriteCloser interface {
	Write(Message) (Message, error)
	Close()
}

// ClientConn represents a client connection to a TCP server.
type ClientConn struct {
	addr    string
	opts    options
	netid   int64
	rawConn net.Conn
	timing  *TimingWheel
	mu      sync.Mutex // guards following
	name    string
	heart   int64
	pending []int64
	once    *sync.Once
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	sending chan bool
}

// NewClientConn returns a new client connection which has not started to
// serve requests yet.
func NewClientConn(netid int64, c net.Conn, opt ...Option) *ClientConn {
	var opts options
	for _, o := range opt {
		o(&opts)
	}
	if opts.bufferSize <= 0 {
		opts.bufferSize = BufferSize256
	}
	return newClientConnWithOptions(netid, c, opts)
}

func newClientConnWithOptions(netid int64, c net.Conn, opts options) *ClientConn {
	cc := &ClientConn{
		addr:    c.RemoteAddr().String(),
		opts:    opts,
		netid:   netid,
		rawConn: c,
		heart:   time.Now().UnixNano(),
		once:    &sync.Once{},
		wg:      &sync.WaitGroup{},
	}
	cc.ctx, cc.cancel = context.WithCancel(context.Background())
	cc.timing = NewTimingWheel(cc.ctx)
	cc.name = c.RemoteAddr().String()
	cc.pending = []int64{}
	cc.sending = make(chan bool, 1)
	return cc
}

// NetID returns the net ID of client connection.
func (cc *ClientConn) NetID() int64 {
	return cc.netid
}

// SetName sets the name of client connection.
func (cc *ClientConn) SetName(name string) {
	cc.mu.Lock()
	cc.name = name
	cc.mu.Unlock()
}

// Name gets the name of client connection.
func (cc *ClientConn) Name() string {
	cc.mu.Lock()
	name := cc.name
	cc.mu.Unlock()
	return name
}

// SetHeartBeat sets the heart beats of client connection.
func (cc *ClientConn) SetHeartBeat(heart int64) {
	cc.mu.Lock()
	cc.heart = heart
	cc.mu.Unlock()
}

// HeartBeat gets the heart beats of client connection.
func (cc *ClientConn) HeartBeat() int64 {
	cc.mu.Lock()
	heart := cc.heart
	cc.mu.Unlock()
	return heart
}

// SetContextValue sets extra data to client connection.
func (cc *ClientConn) SetContextValue(k, v interface{}) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.ctx = context.WithValue(cc.ctx, k, v)
}

// ContextValue gets extra data from client connection.
func (cc *ClientConn) ContextValue(k interface{}) interface{} {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.ctx.Value(k)
}

// Start starts the client connection, creating go-routines for reading,
// writing and handlng.
func (cc *ClientConn) Start() {
	logger.Infof("conn start, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())
	onConnect := cc.opts.onConnect
	if onConnect != nil {
		onConnect(cc)
	}
	loopers := []func(WriteCloser, *sync.WaitGroup){handleLoop}
	for _, l := range loopers {
		looper := l
		cc.wg.Add(1)
		go looper(cc, cc.wg)
	}
}

// Close gracefully closes the client connection. It blocked until all sub
// go-routines are completed and returned.
func (cc *ClientConn) Close() {
	cc.once.Do(func() {
		logger.Infof("conn close gracefully, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())

		// callback on close
		onClose := cc.opts.onClose
		if onClose != nil {
			onClose(cc)
		}

		// close net.Conn, any blocked read or write operation will be unblocked and
		// return errors.
		cc.rawConn.Close()

		// cancel readLoop, writeLoop and handleLoop go-routines.
		cc.mu.Lock()
		cc.cancel()
		cc.pending = nil
		cc.mu.Unlock()

		// stop timer
		cc.timing.Stop()
		cc.wg.Wait()

		if cc.opts.reconnect {
			cc.reconnect()
		}
	})
}

// reconnect reconnects and returns a new *ClientConn.
func (cc *ClientConn) reconnect() {
	c, err := net.Dial("tcp", cc.addr)
	if err != nil {
		logger.Fatalln("net dial error", err)
		return
	}

	*cc = *newClientConnWithOptions(cc.netid, c, cc.opts)
	cc.Start()
}

// Write writes a message to the client.
func (cc *ClientConn) Write(message Message) (Message, error) {
	cc.sending <- true
	pkt, err := cc.opts.codec.Encode(message)
	if _, err = cc.rawConn.Write(pkt); err != nil {
		return nil, err
	}
	<-cc.sending
	return cc.Decode()
}

// RunAt runs a callback at the specified timestamp.
func (cc *ClientConn) RunAt(timestamp time.Time, callback func(time.Time, WriteCloser)) int64 {
	id := runAt(cc.ctx, cc.netid, cc.timing, timestamp, callback)
	if id >= 0 {
		cc.AddPendingTimer(id)
	}
	return id
}

// RunAfter runs a callback right after the specified duration ellapsed.
func (cc *ClientConn) RunAfter(duration time.Duration, callback func(time.Time, WriteCloser)) int64 {
	id := runAfter(cc.ctx, cc.netid, cc.timing, duration, callback)
	if id >= 0 {
		cc.AddPendingTimer(id)
	}
	return id
}

// RunEvery runs a callback on every interval time.
func (cc *ClientConn) RunEvery(interval time.Duration, callback func(time.Time, WriteCloser)) int64 {
	id := runEvery(cc.ctx, cc.netid, cc.timing, interval, callback)
	if id >= 0 {
		cc.AddPendingTimer(id)
	}
	return id
}

// AddPendingTimer adds a new timer ID to client connection.
func (cc *ClientConn) AddPendingTimer(timerID int64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.pending != nil {
		cc.pending = append(cc.pending, timerID)
	}
}

// CancelTimer cancels a timer with the specified ID.
func (cc *ClientConn) CancelTimer(timerID int64) {
	cancelTimer(cc.timing, timerID)
}

// RemoteAddr returns the remote address of server connection.
func (cc *ClientConn) RemoteAddr() net.Addr {
	return cc.rawConn.RemoteAddr()
}

// LocalAddr returns the local address of server connection.
func (cc *ClientConn) LocalAddr() net.Addr {
	return cc.rawConn.LocalAddr()
}

func (cc *ClientConn) Decode() (Message, error) {
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("panics: %v\n", p)
			cc.Close()
		}
	}()

	msg, err := cc.opts.codec.Decode(cc.rawConn)
	if err != nil {
		logger.Errorf("error decoding message %v\n", err)
		if _, ok := err.(ErrUndefined); ok {
			// update heart beats
			cc.SetHeartBeat(time.Now().UnixNano())
		}
		return nil, err
	}
	cc.SetHeartBeat(time.Now().UnixNano())
	return msg, err
}

type contextKey string

// Context keys for messge, server and net ID.
const (
	messageCtx contextKey = "message"
	serverCtx  contextKey = "server"
	netIDCtx   contextKey = "netid"
)

func NewContextWithNetID(ctx context.Context, netID int64) context.Context {
	return context.WithValue(ctx, netIDCtx, netID)
}

func NetIDFromContext(ctx context.Context) int64 {
	return ctx.Value(netIDCtx).(int64)
}

func runAt(ctx context.Context, netID int64, timing *TimingWheel, ts time.Time, cb func(time.Time, WriteCloser)) int64 {
	timeout := NewOnTimeOut(NewContextWithNetID(ctx, netID), cb)
	return timing.AddTimer(ts, 0, timeout)
}

func runAfter(ctx context.Context, netID int64, timing *TimingWheel, d time.Duration, cb func(time.Time, WriteCloser)) int64 {
	delay := time.Now().Add(d)
	return runAt(ctx, netID, timing, delay, cb)
}

func runEvery(ctx context.Context, netID int64, timing *TimingWheel, d time.Duration, cb func(time.Time, WriteCloser)) int64 {
	delay := time.Now().Add(d)
	timeout := NewOnTimeOut(NewContextWithNetID(ctx, netID), cb)
	return timing.AddTimer(delay, d, timeout)
}

func handleLoop(c WriteCloser, wg *sync.WaitGroup) {
	var (
		timerCh chan *OnTimeOut
		netID   int64
	)

	switch c := c.(type) {
	case *ClientConn:
		timerCh = c.timing.timeOutChan
		netID = c.netid
	}

	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("panics: %v\n", p)
		}
		wg.Done()
		logger.Debugln("handleLoop go-routine exited")
		c.Close()
	}()

	for {
		select {
		case timeout := <-timerCh:
			if timeout != nil {
				timeoutNetID := NetIDFromContext(timeout.Ctx)
				if timeoutNetID != netID {
					logger.Errorf("timeout net %d, conn net %d, mismatched!\n", timeoutNetID, netID)
				}
				timeout.Callback(time.Now(), c.(WriteCloser))
			}
		}
	}
}
