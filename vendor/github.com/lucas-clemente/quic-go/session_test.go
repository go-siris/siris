package quic

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"runtime/pprof"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/lucas-clemente/quic-go/ackhandler"
	"github.com/lucas-clemente/quic-go/crypto"
	"github.com/lucas-clemente/quic-go/frames"
	"github.com/lucas-clemente/quic-go/handshake"
	"github.com/lucas-clemente/quic-go/internal/mocks"
	"github.com/lucas-clemente/quic-go/internal/mocks/mocks_fc"
	"github.com/lucas-clemente/quic-go/protocol"
	"github.com/lucas-clemente/quic-go/qerr"
	"github.com/lucas-clemente/quic-go/testdata"
)

type mockConnection struct {
	remoteAddr net.Addr
	localAddr  net.Addr
	written    [][]byte
}

func (m *mockConnection) Write(p []byte) error {
	b := make([]byte, len(p))
	copy(b, p)
	m.written = append(m.written, b)
	return nil
}
func (m *mockConnection) Read([]byte) (int, net.Addr, error) { panic("not implemented") }

func (m *mockConnection) SetCurrentRemoteAddr(addr net.Addr) {
	m.remoteAddr = addr
}
func (m *mockConnection) LocalAddr() net.Addr  { return m.localAddr }
func (m *mockConnection) RemoteAddr() net.Addr { return m.remoteAddr }
func (*mockConnection) Close() error           { panic("not implemented") }

type mockUnpacker struct {
	unpackErr error
}

func (m *mockUnpacker) Unpack(publicHeaderBinary []byte, hdr *PublicHeader, data []byte) (*unpackedPacket, error) {
	if m.unpackErr != nil {
		return nil, m.unpackErr
	}
	return &unpackedPacket{
		frames: nil,
	}, nil
}

type mockSentPacketHandler struct {
	retransmissionQueue  []*ackhandler.Packet
	sentPackets          []*ackhandler.Packet
	congestionLimited    bool
	requestedStopWaiting bool
}

func (h *mockSentPacketHandler) SentPacket(packet *ackhandler.Packet) error {
	h.sentPackets = append(h.sentPackets, packet)
	return nil
}

func (h *mockSentPacketHandler) ReceivedAck(ackFrame *frames.AckFrame, withPacketNumber protocol.PacketNumber, recvTime time.Time) error {
	return nil
}

func (h *mockSentPacketHandler) GetLeastUnacked() protocol.PacketNumber { return 1 }
func (h *mockSentPacketHandler) GetAlarmTimeout() time.Time             { panic("not implemented") }
func (h *mockSentPacketHandler) OnAlarm()                               { panic("not implemented") }
func (h *mockSentPacketHandler) SendingAllowed() bool                   { return !h.congestionLimited }

func (h *mockSentPacketHandler) GetStopWaitingFrame(force bool) *frames.StopWaitingFrame {
	h.requestedStopWaiting = true
	return &frames.StopWaitingFrame{LeastUnacked: 0x1337}
}

func (h *mockSentPacketHandler) DequeuePacketForRetransmission() *ackhandler.Packet {
	if len(h.retransmissionQueue) > 0 {
		packet := h.retransmissionQueue[0]
		h.retransmissionQueue = h.retransmissionQueue[1:]
		return packet
	}
	return nil
}

func newMockSentPacketHandler() ackhandler.SentPacketHandler {
	return &mockSentPacketHandler{}
}

var _ ackhandler.SentPacketHandler = &mockSentPacketHandler{}

type mockReceivedPacketHandler struct {
	nextAckFrame *frames.AckFrame
	ackAlarm     time.Time
}

func (m *mockReceivedPacketHandler) GetAckFrame() *frames.AckFrame {
	f := m.nextAckFrame
	m.nextAckFrame = nil
	return f
}
func (m *mockReceivedPacketHandler) ReceivedPacket(packetNumber protocol.PacketNumber, shouldInstigateAck bool) error {
	panic("not implemented")
}
func (m *mockReceivedPacketHandler) ReceivedStopWaiting(*frames.StopWaitingFrame) error {
	panic("not implemented")
}
func (m *mockReceivedPacketHandler) GetAlarmTimeout() time.Time { return m.ackAlarm }

var _ ackhandler.ReceivedPacketHandler = &mockReceivedPacketHandler{}

func areSessionsRunning() bool {
	var b bytes.Buffer
	pprof.Lookup("goroutine").WriteTo(&b, 1)
	return strings.Contains(b.String(), "quic-go.(*session).run")
}

var _ = Describe("Session", func() {
	var (
		sess          *session
		scfg          *handshake.ServerConfig
		mconn         *mockConnection
		mockCpm       *mocks.MockConnectionParametersManager
		cryptoSetup   *mockCryptoSetup
		handshakeChan <-chan handshakeEvent
		aeadChanged   chan<- protocol.EncryptionLevel
	)

	BeforeEach(func() {
		Eventually(areSessionsRunning).Should(BeFalse())

		cryptoSetup = &mockCryptoSetup{}
		newCryptoSetup = func(
			_ protocol.ConnectionID,
			_ net.Addr,
			_ protocol.VersionNumber,
			_ *handshake.ServerConfig,
			_ io.ReadWriter,
			_ handshake.ConnectionParametersManager,
			_ []protocol.VersionNumber,
			_ func(net.Addr, *handshake.STK) bool,
			aeadChangedP chan<- protocol.EncryptionLevel,
		) (handshake.CryptoSetup, error) {
			aeadChanged = aeadChangedP
			return cryptoSetup, nil
		}

		mconn = &mockConnection{
			remoteAddr: &net.UDPAddr{},
		}
		certChain := crypto.NewCertChain(testdata.GetTLSConfig())
		kex, err := crypto.NewCurve25519KEX()
		Expect(err).NotTo(HaveOccurred())
		scfg, err = handshake.NewServerConfig(kex, certChain)
		Expect(err).NotTo(HaveOccurred())
		var pSess Session
		pSess, handshakeChan, err = newSession(
			mconn,
			protocol.Version35,
			0,
			scfg,
			nil,
			populateServerConfig(&Config{}),
		)
		Expect(err).NotTo(HaveOccurred())
		sess = pSess.(*session)
		Expect(sess.streamsMap.openStreams).To(HaveLen(1)) // Crypto stream

		mockCpm = mocks.NewMockConnectionParametersManager(mockCtrl)
		mockCpm.EXPECT().GetIdleConnectionStateLifetime().Return(time.Minute).AnyTimes()
		sess.connectionParameters = mockCpm
	})

	AfterEach(func() {
		newCryptoSetup = handshake.NewCryptoSetup
		Eventually(areSessionsRunning).Should(BeFalse())
	})

	Context("source address validation", func() {
		var (
			stkVerify       func(net.Addr, *handshake.STK) bool
			paramClientAddr net.Addr
			paramSTK        *STK
		)
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 13, 37), Port: 1000}

		BeforeEach(func() {
			newCryptoSetup = func(
				_ protocol.ConnectionID,
				_ net.Addr,
				_ protocol.VersionNumber,
				_ *handshake.ServerConfig,
				_ io.ReadWriter,
				_ handshake.ConnectionParametersManager,
				_ []protocol.VersionNumber,
				stkFunc func(net.Addr, *handshake.STK) bool,
				_ chan<- protocol.EncryptionLevel,
			) (handshake.CryptoSetup, error) {
				stkVerify = stkFunc
				return cryptoSetup, nil
			}

			conf := populateServerConfig(&Config{})
			conf.AcceptSTK = func(clientAddr net.Addr, stk *STK) bool {
				paramClientAddr = clientAddr
				paramSTK = stk
				return false
			}
			pSess, _, err := newSession(
				mconn,
				protocol.Version35,
				0,
				scfg,
				nil,
				conf,
			)
			Expect(err).NotTo(HaveOccurred())
			sess = pSess.(*session)
		})

		It("calls the callback with the right parameters when the client didn't send an STK", func() {
			stkVerify(remoteAddr, nil)
			Expect(paramClientAddr).To(Equal(remoteAddr))
			Expect(paramSTK).To(BeNil())
		})

		It("calls the callback with the STK when the client sent an STK", func() {
			stkAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1337}
			sentTime := time.Now().Add(-time.Hour)
			stkVerify(remoteAddr, &handshake.STK{SentTime: sentTime, RemoteAddr: stkAddr.String()})
			Expect(paramClientAddr).To(Equal(remoteAddr))
			Expect(paramSTK).ToNot(BeNil())
			Expect(paramSTK.remoteAddr).To(Equal(stkAddr.String()))
			Expect(paramSTK.sentTime).To(Equal(sentTime))
		})
	})

	Context("when handling stream frames", func() {
		It("makes new streams", func() {
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte{0xde, 0xca, 0xfb, 0xad},
			})
			p := make([]byte, 4)
			str, err := sess.streamsMap.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			_, err = str.Read(p)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(Equal([]byte{0xde, 0xca, 0xfb, 0xad}))
		})

		It("does not reject existing streams with even StreamIDs", func() {
			_, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			err = sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte{0xde, 0xca, 0xfb, 0xad},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("handles existing streams", func() {
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte{0xde, 0xca},
			})
			numOpenStreams := len(sess.streamsMap.openStreams)
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Offset:   2,
				Data:     []byte{0xfb, 0xad},
			})
			Expect(sess.streamsMap.openStreams).To(HaveLen(numOpenStreams))
			p := make([]byte, 4)
			str, _ := sess.streamsMap.GetOrOpenStream(5)
			Expect(str).ToNot(BeNil())
			_, err := str.Read(p)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(Equal([]byte{0xde, 0xca, 0xfb, 0xad}))
		})

		It("does not delete streams with Close()", func() {
			str, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			str.Close()
			sess.garbageCollectStreams()
			str, err = sess.streamsMap.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		It("does not delete streams with FIN bit", func() {
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte{0xde, 0xca, 0xfb, 0xad},
				FinBit:   true,
			})
			numOpenStreams := len(sess.streamsMap.openStreams)
			str, _ := sess.streamsMap.GetOrOpenStream(5)
			Expect(str).ToNot(BeNil())
			p := make([]byte, 4)
			_, err := str.Read(p)
			Expect(err).To(MatchError(io.EOF))
			Expect(p).To(Equal([]byte{0xde, 0xca, 0xfb, 0xad}))
			sess.garbageCollectStreams()
			Expect(sess.streamsMap.openStreams).To(HaveLen(numOpenStreams))
			str, _ = sess.streamsMap.GetOrOpenStream(5)
			Expect(str).ToNot(BeNil())
		})

		It("deletes streams with FIN bit & close", func() {
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte{0xde, 0xca, 0xfb, 0xad},
				FinBit:   true,
			})
			numOpenStreams := len(sess.streamsMap.openStreams)
			str, _ := sess.streamsMap.GetOrOpenStream(5)
			Expect(str).ToNot(BeNil())
			p := make([]byte, 4)
			_, err := str.Read(p)
			Expect(err).To(MatchError(io.EOF))
			Expect(p).To(Equal([]byte{0xde, 0xca, 0xfb, 0xad}))
			sess.garbageCollectStreams()
			Expect(sess.streamsMap.openStreams).To(HaveLen(numOpenStreams))
			str, _ = sess.streamsMap.GetOrOpenStream(5)
			Expect(str).ToNot(BeNil())
			// We still need to close the stream locally
			str.Close()
			// ... and simulate that we actually the FIN
			str.sentFin()
			sess.garbageCollectStreams()
			Expect(len(sess.streamsMap.openStreams)).To(BeNumerically("<", numOpenStreams))
			str, err = sess.streamsMap.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).To(BeNil())
			// flow controller should have been notified
			_, err = sess.flowControlManager.SendWindowSize(5)
			Expect(err).To(MatchError("Error accessing the flowController map."))
		})

		It("cancels streams with error", func() {
			sess.garbageCollectStreams()
			testErr := errors.New("test")
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte{0xde, 0xca, 0xfb, 0xad},
			})
			str, err := sess.streamsMap.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			p := make([]byte, 4)
			_, err = str.Read(p)
			Expect(err).ToNot(HaveOccurred())
			sess.handleCloseError(closeError{err: testErr, remote: true})
			_, err = str.Read(p)
			Expect(err).To(MatchError(qerr.Error(qerr.InternalError, testErr.Error())))
			sess.garbageCollectStreams()
			str, err = sess.streamsMap.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("cancels empty streams with error", func() {
			testErr := errors.New("test")
			sess.GetOrOpenStream(5)
			str, err := sess.streamsMap.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			sess.handleCloseError(closeError{err: testErr, remote: true})
			_, err = str.Read([]byte{0})
			Expect(err).To(MatchError(qerr.Error(qerr.InternalError, testErr.Error())))
			sess.garbageCollectStreams()
			str, err = sess.streamsMap.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("informs the FlowControlManager about new streams", func() {
			// since the stream doesn't yet exist, this will throw an error
			err := sess.flowControlManager.UpdateHighestReceived(5, 1000)
			Expect(err).To(HaveOccurred())
			sess.GetOrOpenStream(5)
			err = sess.flowControlManager.UpdateHighestReceived(5, 2000)
			Expect(err).ToNot(HaveOccurred())
		})

		It("ignores STREAM frames for closed streams (client-side)", func() {
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				FinBit:   true,
			})
			str, _ := sess.streamsMap.GetOrOpenStream(5)
			Expect(str).ToNot(BeNil())
			_, err := str.Read([]byte{0})
			Expect(err).To(MatchError(io.EOF))
			str.Close()
			str.sentFin()
			sess.garbageCollectStreams()
			str, _ = sess.streamsMap.GetOrOpenStream(5)
			Expect(str).To(BeNil()) // make sure the stream is gone
			err = sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte("foobar"),
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("ignores STREAM frames for closed streams (server-side)", func() {
			ostr, err := sess.OpenStream()
			Expect(err).ToNot(HaveOccurred())
			Expect(ostr.StreamID()).To(Equal(protocol.StreamID(2)))
			err = sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 2,
				FinBit:   true,
			})
			Expect(err).ToNot(HaveOccurred())
			str, _ := sess.streamsMap.GetOrOpenStream(2)
			Expect(str).ToNot(BeNil())
			_, err = str.Read([]byte{0})
			Expect(err).To(MatchError(io.EOF))
			str.Close()
			str.sentFin()
			sess.garbageCollectStreams()
			str, _ = sess.streamsMap.GetOrOpenStream(2)
			Expect(str).To(BeNil()) // make sure the stream is gone
			err = sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 2,
				FinBit:   true,
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("handling RST_STREAM frames", func() {
		It("closes the streams for writing", func() {
			s, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			err = sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID:  5,
				ErrorCode: 42,
			})
			Expect(err).ToNot(HaveOccurred())
			n, err := s.Write([]byte{0})
			Expect(n).To(BeZero())
			Expect(err).To(MatchError("RST_STREAM received with code 42"))
		})

		It("doesn't close the stream for reading", func() {
			s, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 5,
				Data:     []byte("foobar"),
			})
			err = sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID:   5,
				ErrorCode:  42,
				ByteOffset: 6,
			})
			Expect(err).ToNot(HaveOccurred())
			b := make([]byte, 3)
			n, err := s.Read(b)
			Expect(n).To(Equal(3))
			Expect(err).ToNot(HaveOccurred())
		})

		It("queues a RST_STERAM frame with the correct offset", func() {
			str, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			str.(*stream).writeOffset = 0x1337
			err = sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID: 5,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.packer.controlFrames).To(HaveLen(1))
			Expect(sess.packer.controlFrames[0].(*frames.RstStreamFrame)).To(Equal(&frames.RstStreamFrame{
				StreamID:   5,
				ByteOffset: 0x1337,
			}))
			Expect(str.(*stream).finished()).To(BeTrue())
		})

		It("doesn't queue a RST_STREAM for a stream that it already sent a FIN on", func() {
			str, err := sess.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			str.(*stream).sentFin()
			str.Close()
			err = sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID: 5,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.packer.controlFrames).To(BeEmpty())
			Expect(str.(*stream).finished()).To(BeTrue())
		})

		It("passes the byte offset to the flow controller", func() {
			sess.streamsMap.GetOrOpenStream(5)
			fcm := mocks_fc.NewMockFlowControlManager(mockCtrl)
			sess.flowControlManager = fcm
			fcm.EXPECT().ResetStream(protocol.StreamID(5), protocol.ByteCount(0x1337))
			err := sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID:   5,
				ByteOffset: 0x1337,
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns errors from the flow controller", func() {
			testErr := errors.New("flow control violation")
			sess.streamsMap.GetOrOpenStream(5)
			fcm := mocks_fc.NewMockFlowControlManager(mockCtrl)
			sess.flowControlManager = fcm
			fcm.EXPECT().ResetStream(protocol.StreamID(5), protocol.ByteCount(0x1337)).Return(testErr)
			err := sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID:   5,
				ByteOffset: 0x1337,
			})
			Expect(err).To(MatchError(testErr))
		})

		It("ignores the error when the stream is not known", func() {
			err := sess.handleFrames([]frames.Frame{&frames.RstStreamFrame{
				StreamID:  5,
				ErrorCode: 42,
			}})
			Expect(err).NotTo(HaveOccurred())
		})

		It("queues a RST_STREAM when a stream gets reset locally", func() {
			testErr := errors.New("testErr")
			str, err := sess.streamsMap.GetOrOpenStream(5)
			str.writeOffset = 0x1337
			Expect(err).ToNot(HaveOccurred())
			str.Reset(testErr)
			Expect(sess.packer.controlFrames).To(HaveLen(1))
			Expect(sess.packer.controlFrames[0]).To(Equal(&frames.RstStreamFrame{
				StreamID:   5,
				ByteOffset: 0x1337,
			}))
			Expect(str.finished()).To(BeFalse())
		})

		It("doesn't queue another RST_STREAM, when it receives an RST_STREAM as a response for the first", func() {
			testErr := errors.New("testErr")
			str, err := sess.streamsMap.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			str.Reset(testErr)
			Expect(sess.packer.controlFrames).To(HaveLen(1))
			err = sess.handleRstStreamFrame(&frames.RstStreamFrame{
				StreamID:   5,
				ByteOffset: 0x42,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.packer.controlFrames).To(HaveLen(1))
		})
	})

	Context("handling WINDOW_UPDATE frames", func() {
		It("updates the Flow Control Window of a stream", func() {
			_, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			err = sess.handleWindowUpdateFrame(&frames.WindowUpdateFrame{
				StreamID:   5,
				ByteOffset: 100,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.flowControlManager.SendWindowSize(5)).To(Equal(protocol.ByteCount(100)))
		})

		It("updates the Flow Control Window of the connection", func() {
			err := sess.handleWindowUpdateFrame(&frames.WindowUpdateFrame{
				StreamID:   0,
				ByteOffset: 0x800000,
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("opens a new stream when receiving a WINDOW_UPDATE for an unknown stream", func() {
			err := sess.handleWindowUpdateFrame(&frames.WindowUpdateFrame{
				StreamID:   5,
				ByteOffset: 1337,
			})
			Expect(err).ToNot(HaveOccurred())
			str, err := sess.streamsMap.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		It("errors when receiving a WindowUpdateFrame for a closed stream", func() {
			sess.handleStreamFrame(&frames.StreamFrame{StreamID: 5})
			err := sess.streamsMap.RemoveStream(5)
			Expect(err).ToNot(HaveOccurred())
			sess.garbageCollectStreams()
			err = sess.handleWindowUpdateFrame(&frames.WindowUpdateFrame{
				StreamID:   5,
				ByteOffset: 1337,
			})
			Expect(err).To(MatchError(errWindowUpdateOnClosedStream))
		})

		It("ignores errors when receiving a WindowUpdateFrame for a closed stream", func() {
			sess.handleStreamFrame(&frames.StreamFrame{StreamID: 5})
			err := sess.streamsMap.RemoveStream(5)
			Expect(err).ToNot(HaveOccurred())
			sess.garbageCollectStreams()
			err = sess.handleFrames([]frames.Frame{&frames.WindowUpdateFrame{
				StreamID:   5,
				ByteOffset: 1337,
			}})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("handles PING frames", func() {
		err := sess.handleFrames([]frames.Frame{&frames.PingFrame{}})
		Expect(err).NotTo(HaveOccurred())
	})

	It("handles BLOCKED frames", func() {
		err := sess.handleFrames([]frames.Frame{&frames.BlockedFrame{}})
		Expect(err).NotTo(HaveOccurred())
	})

	It("errors on GOAWAY frames", func() {
		err := sess.handleFrames([]frames.Frame{&frames.GoawayFrame{}})
		Expect(err).To(MatchError("unimplemented: handling GOAWAY frames"))
	})

	It("handles STOP_WAITING frames", func() {
		err := sess.handleFrames([]frames.Frame{&frames.StopWaitingFrame{LeastUnacked: 10}})
		Expect(err).NotTo(HaveOccurred())
	})

	It("handles CONNECTION_CLOSE frames", func(done Done) {
		go sess.run()
		str, _ := sess.GetOrOpenStream(5)
		err := sess.handleFrames([]frames.Frame{&frames.ConnectionCloseFrame{ErrorCode: 42, ReasonPhrase: "foobar"}})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess.runClosed).Should(BeClosed())
		_, err = str.Read([]byte{0})
		Expect(err).To(MatchError(qerr.Error(42, "foobar")))
		close(done)
	})

	Context("waiting until the handshake completes", func() {
		It("waits until the handshake is complete", func(done Done) {
			go sess.run()

			var waitReturned bool
			go func() {
				defer GinkgoRecover()
				err := sess.WaitUntilHandshakeComplete()
				Expect(err).ToNot(HaveOccurred())
				waitReturned = true
			}()
			aeadChanged <- protocol.EncryptionForwardSecure
			Consistently(func() bool { return waitReturned }).Should(BeFalse())
			close(aeadChanged)
			Eventually(func() bool { return waitReturned }).Should(BeTrue())
			Expect(sess.Close(nil)).To(Succeed())
			close(done)
		})

		It("errors if the handshake fails", func(done Done) {
			testErr := errors.New("crypto error")
			sess.cryptoSetup = &mockCryptoSetup{handleErr: testErr}
			go sess.run()
			err := sess.WaitUntilHandshakeComplete()
			Expect(err).To(MatchError(testErr))
			close(done)
		}, 0.5)

		It("returns when Close is called", func(done Done) {
			testErr := errors.New("close error")
			go sess.run()
			var waitReturned bool
			go func() {
				defer GinkgoRecover()
				err := sess.WaitUntilHandshakeComplete()
				Expect(err).To(MatchError(testErr))
				waitReturned = true
			}()
			sess.Close(testErr)
			Eventually(func() bool { return waitReturned }).Should(BeTrue())
			close(done)
		})

		It("doesn't wait if the handshake is already completed", func(done Done) {
			go sess.run()
			close(aeadChanged)
			err := sess.WaitUntilHandshakeComplete()
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.Close(nil)).To(Succeed())
			close(done)
		})
	})

	Context("accepting streams", func() {
		It("waits for new streams", func() {
			var str Stream
			go func() {
				defer GinkgoRecover()
				var err error
				str, err = sess.AcceptStream()
				Expect(err).ToNot(HaveOccurred())
			}()
			Consistently(func() Stream { return str }).Should(BeNil())
			sess.handleStreamFrame(&frames.StreamFrame{
				StreamID: 3,
			})
			Eventually(func() Stream { return str }).ShouldNot(BeNil())
			Expect(str.StreamID()).To(Equal(protocol.StreamID(3)))
		})

		It("stops accepting when the session is closed", func() {
			testErr := errors.New("testErr")
			var err error
			go func() {
				_, err = sess.AcceptStream()
			}()
			go sess.run()
			Consistently(func() error { return err }).ShouldNot(HaveOccurred())
			sess.Close(testErr)
			Eventually(func() error { return err }).Should(HaveOccurred())
			Expect(err).To(MatchError(qerr.ToQuicError(testErr)))
		})

		It("stops accepting when the session is closed after version negotiation", func() {
			var err error
			go func() {
				_, err = sess.AcceptStream()
			}()
			go sess.run()
			Consistently(func() error { return err }).ShouldNot(HaveOccurred())
			Expect(sess.runClosed).ToNot(BeClosed())
			sess.Close(errCloseSessionForNewVersion)
			Eventually(func() error { return err }).Should(HaveOccurred())
			Expect(err).To(MatchError(qerr.Error(qerr.InternalError, errCloseSessionForNewVersion.Error())))
			Eventually(sess.runClosed).Should(BeClosed())
		})
	})

	Context("closing", func() {
		BeforeEach(func() {
			Eventually(areSessionsRunning).Should(BeFalse())
			go sess.run()
			Eventually(areSessionsRunning).Should(BeTrue())
		})

		It("shuts down without error", func() {
			sess.Close(nil)
			Eventually(areSessionsRunning).Should(BeFalse())
			Expect(mconn.written).To(HaveLen(1))
			Expect(mconn.written[0]).To(ContainSubstring(string([]byte{0x02, byte(qerr.PeerGoingAway), 0, 0, 0, 0, 0})))
			Expect(sess.runClosed).To(BeClosed())
		})

		It("only closes once", func() {
			sess.Close(nil)
			sess.Close(nil)
			Eventually(areSessionsRunning).Should(BeFalse())
			Expect(mconn.written).To(HaveLen(1))
			Expect(sess.runClosed).To(BeClosed())
		})

		It("closes streams with proper error", func() {
			testErr := errors.New("test error")
			s, err := sess.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			sess.Close(testErr)
			Eventually(areSessionsRunning).Should(BeFalse())
			n, err := s.Read([]byte{0})
			Expect(n).To(BeZero())
			Expect(err.Error()).To(ContainSubstring(testErr.Error()))
			n, err = s.Write([]byte{0})
			Expect(n).To(BeZero())
			Expect(err.Error()).To(ContainSubstring(testErr.Error()))
			Expect(sess.runClosed).To(BeClosed())
		})

		It("closes the session in order to replace it with another QUIC version", func() {
			sess.Close(errCloseSessionForNewVersion)
			Eventually(areSessionsRunning).Should(BeFalse())
			Expect(mconn.written).To(BeEmpty()) // no CONNECTION_CLOSE or PUBLIC_RESET sent
		})

		It("sends a Public Reset if the client is initiating the head-of-line blocking experiment", func() {
			sess.Close(handshake.ErrHOLExperiment)
			Expect(mconn.written).To(HaveLen(1))
			Expect(mconn.written[0][0] & 0x02).ToNot(BeZero()) // Public Reset
			Expect(sess.runClosed).To(BeClosed())
		})

		It("unblocks WaitUntilClosed when the run loop exists", func() {
			returned := make(chan struct{})
			go func() {
				sess.WaitUntilClosed()
				close(returned)
			}()
			Consistently(returned).ShouldNot(BeClosed())
			sess.Close(nil)
			Eventually(returned).Should(BeClosed())
		})
	})

	Context("receiving packets", func() {
		var hdr *PublicHeader

		BeforeEach(func() {
			sess.unpacker = &mockUnpacker{}
			hdr = &PublicHeader{PacketNumberLen: protocol.PacketNumberLen6}
		})

		It("sets the {last,largest}RcvdPacketNumber", func() {
			hdr.PacketNumber = 5
			err := sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.lastRcvdPacketNumber).To(Equal(protocol.PacketNumber(5)))
			Expect(sess.largestRcvdPacketNumber).To(Equal(protocol.PacketNumber(5)))
		})

		It("closes when handling a packet fails", func(done Done) {
			testErr := errors.New("unpack error")
			hdr.PacketNumber = 5
			var runErr error
			go func() {
				runErr = sess.run()
			}()
			sess.unpacker.(*mockUnpacker).unpackErr = testErr
			sess.handlePacket(&receivedPacket{publicHeader: hdr})
			Eventually(func() error { return runErr }).Should(MatchError(testErr))
			Expect(sess.runClosed).To(BeClosed())
			close(done)
		})

		It("sets the {last,largest}RcvdPacketNumber, for an out-of-order packet", func() {
			hdr.PacketNumber = 5
			err := sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.lastRcvdPacketNumber).To(Equal(protocol.PacketNumber(5)))
			Expect(sess.largestRcvdPacketNumber).To(Equal(protocol.PacketNumber(5)))
			hdr.PacketNumber = 3
			err = sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
			Expect(sess.lastRcvdPacketNumber).To(Equal(protocol.PacketNumber(3)))
			Expect(sess.largestRcvdPacketNumber).To(Equal(protocol.PacketNumber(5)))
		})

		It("handles duplicate packets", func() {
			hdr.PacketNumber = 5
			err := sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
			err = sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
		})

		It("handles packets smaller than the highest LeastUnacked of a StopWaiting", func() {
			err := sess.receivedPacketHandler.ReceivedStopWaiting(&frames.StopWaitingFrame{LeastUnacked: 10})
			Expect(err).ToNot(HaveOccurred())
			hdr.PacketNumber = 5
			err = sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("updating the remote address", func() {
			It("sets the remote address", func() {
				remoteIP := &net.IPAddr{IP: net.IPv4(192, 168, 0, 100)}
				Expect(sess.conn.(*mockConnection).remoteAddr).ToNot(Equal(remoteIP))
				p := receivedPacket{
					remoteAddr:   remoteIP,
					publicHeader: &PublicHeader{PacketNumber: 1337},
				}
				err := sess.handlePacketImpl(&p)
				Expect(err).ToNot(HaveOccurred())
				Expect(sess.conn.(*mockConnection).remoteAddr).To(Equal(remoteIP))
			})

			It("doesn't change the remote address if authenticating the packet fails", func() {
				remoteIP := &net.IPAddr{IP: net.IPv4(192, 168, 0, 100)}
				attackerIP := &net.IPAddr{IP: net.IPv4(192, 168, 0, 102)}
				sess.conn.(*mockConnection).remoteAddr = remoteIP
				// use the real packetUnpacker here, to make sure this test fails if the error code for failed decryption changes
				sess.unpacker = &packetUnpacker{}
				sess.unpacker.(*packetUnpacker).aead = &mockAEAD{}
				p := receivedPacket{
					remoteAddr:   attackerIP,
					publicHeader: &PublicHeader{PacketNumber: 1337},
				}
				err := sess.handlePacketImpl(&p)
				quicErr := err.(*qerr.QuicError)
				Expect(quicErr.ErrorCode).To(Equal(qerr.DecryptionFailure))
				Expect(sess.conn.(*mockConnection).remoteAddr).To(Equal(remoteIP))
			})

			It("sets the remote address, if the packet is authenticated, but unpacking fails for another reason", func() {
				testErr := errors.New("testErr")
				remoteIP := &net.IPAddr{IP: net.IPv4(192, 168, 0, 100)}
				Expect(sess.conn.(*mockConnection).remoteAddr).ToNot(Equal(remoteIP))
				p := receivedPacket{
					remoteAddr:   remoteIP,
					publicHeader: &PublicHeader{PacketNumber: 1337},
				}
				sess.unpacker.(*mockUnpacker).unpackErr = testErr
				err := sess.handlePacketImpl(&p)
				Expect(err).To(MatchError(testErr))
				Expect(sess.conn.(*mockConnection).remoteAddr).To(Equal(remoteIP))
			})
		})
	})

	Context("sending packets", func() {
		It("sends ack frames", func() {
			packetNumber := protocol.PacketNumber(0x035E)
			sess.receivedPacketHandler.ReceivedPacket(packetNumber, true)
			err := sess.sendPacket()
			Expect(err).NotTo(HaveOccurred())
			Expect(mconn.written).To(HaveLen(1))
			Expect(mconn.written[0]).To(ContainSubstring(string([]byte{0x5E, 0x03})))
		})

		It("sends ACK frames when congestion limited", func() {
			sess.sentPacketHandler = &mockSentPacketHandler{congestionLimited: true}
			sess.packer.packetNumberGenerator.next = 0x1338
			packetNumber := protocol.PacketNumber(0x035E)
			sess.receivedPacketHandler.ReceivedPacket(packetNumber, true)
			err := sess.sendPacket()
			Expect(err).NotTo(HaveOccurred())
			Expect(mconn.written).To(HaveLen(1))
			Expect(mconn.written[0]).To(ContainSubstring(string([]byte{0x5E, 0x03})))
		})

		It("sends two WindowUpdate frames", func() {
			_, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			sess.flowControlManager.AddBytesRead(5, protocol.ReceiveStreamFlowControlWindow)
			err = sess.sendPacket()
			Expect(err).NotTo(HaveOccurred())
			err = sess.sendPacket()
			Expect(err).NotTo(HaveOccurred())
			err = sess.sendPacket()
			Expect(err).NotTo(HaveOccurred())
			Expect(mconn.written).To(HaveLen(2))
			Expect(mconn.written[0]).To(ContainSubstring(string([]byte{0x04, 0x05, 0, 0, 0})))
			Expect(mconn.written[1]).To(ContainSubstring(string([]byte{0x04, 0x05, 0, 0, 0})))
		})

		It("sends public reset", func() {
			err := sess.sendPublicReset(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(mconn.written).To(HaveLen(1))
			Expect(mconn.written[0]).To(ContainSubstring("PRST"))
		})

		It("informs the SentPacketHandler about sent packets", func() {
			sess.sentPacketHandler = newMockSentPacketHandler()
			sess.packer.packetNumberGenerator.next = 0x1337 + 9
			sess.packer.cryptoSetup = &mockCryptoSetup{encLevelSeal: protocol.EncryptionForwardSecure}

			f := &frames.StreamFrame{
				StreamID: 5,
				Data:     []byte("foobar"),
			}
			sess.streamFramer.AddFrameForRetransmission(f)
			_, err := sess.GetOrOpenStream(5)
			Expect(err).ToNot(HaveOccurred())
			err = sess.sendPacket()
			Expect(err).NotTo(HaveOccurred())
			Expect(mconn.written).To(HaveLen(1))
			sentPackets := sess.sentPacketHandler.(*mockSentPacketHandler).sentPackets
			Expect(sentPackets).To(HaveLen(1))
			Expect(sentPackets[0].Frames).To(ContainElement(f))
			Expect(sentPackets[0].EncryptionLevel).To(Equal(protocol.EncryptionForwardSecure))
			Expect(sentPackets[0].Length).To(BeEquivalentTo(len(mconn.written[0])))
		})
	})

	Context("retransmissions", func() {
		var sph *mockSentPacketHandler
		BeforeEach(func() {
			// a StopWaitingFrame is added, so make sure the packet number of the new package is higher than the packet number of the retransmitted packet
			sess.packer.packetNumberGenerator.next = 0x1337 + 10
			sph = newMockSentPacketHandler().(*mockSentPacketHandler)
			sess.sentPacketHandler = sph
			sess.packer.cryptoSetup = &mockCryptoSetup{encLevelSeal: protocol.EncryptionForwardSecure}
		})

		Context("for handshake packets", func() {
			It("retransmits an unencrypted packet", func() {
				sf := &frames.StreamFrame{StreamID: 1, Data: []byte("foobar")}
				sph.retransmissionQueue = []*ackhandler.Packet{{
					Frames:          []frames.Frame{sf},
					EncryptionLevel: protocol.EncryptionUnencrypted,
				}}
				err := sess.sendPacket()
				Expect(err).ToNot(HaveOccurred())
				Expect(mconn.written).To(HaveLen(1))
				sentPackets := sph.sentPackets
				Expect(sentPackets).To(HaveLen(1))
				Expect(sentPackets[0].EncryptionLevel).To(Equal(protocol.EncryptionUnencrypted))
				Expect(sentPackets[0].Frames).To(HaveLen(2))
				Expect(sentPackets[0].Frames[1]).To(Equal(sf))
				swf := sentPackets[0].Frames[0].(*frames.StopWaitingFrame)
				Expect(swf.LeastUnacked).To(Equal(protocol.PacketNumber(0x1337)))
			})

			It("retransmit a packet encrypted with the initial encryption", func() {
				sf := &frames.StreamFrame{StreamID: 1, Data: []byte("foobar")}
				sph.retransmissionQueue = []*ackhandler.Packet{{
					Frames:          []frames.Frame{sf},
					EncryptionLevel: protocol.EncryptionSecure,
				}}
				err := sess.sendPacket()
				Expect(err).ToNot(HaveOccurred())
				Expect(mconn.written).To(HaveLen(1))
				sentPackets := sph.sentPackets
				Expect(sentPackets).To(HaveLen(1))
				Expect(sentPackets[0].EncryptionLevel).To(Equal(protocol.EncryptionSecure))
				Expect(sentPackets[0].Frames).To(HaveLen(2))
				Expect(sentPackets[0].Frames).To(ContainElement(sf))
			})

			It("doesn't retransmit handshake packets when the handshake is complete", func() {
				sess.handshakeComplete = true
				sf := &frames.StreamFrame{StreamID: 1, Data: []byte("foobar")}
				sph.retransmissionQueue = []*ackhandler.Packet{{
					Frames:          []frames.Frame{sf},
					EncryptionLevel: protocol.EncryptionSecure,
				}}
				err := sess.sendPacket()
				Expect(err).ToNot(HaveOccurred())
				Expect(mconn.written).To(BeEmpty())
			})
		})

		Context("for packets after the handshake", func() {
			It("sends a StreamFrame from a packet queued for retransmission", func() {
				f := frames.StreamFrame{
					StreamID: 0x5,
					Data:     []byte("foobar1234567"),
				}
				p := ackhandler.Packet{
					PacketNumber:    0x1337,
					Frames:          []frames.Frame{&f},
					EncryptionLevel: protocol.EncryptionForwardSecure,
				}
				sph.retransmissionQueue = []*ackhandler.Packet{&p}

				err := sess.sendPacket()
				Expect(err).NotTo(HaveOccurred())
				Expect(mconn.written).To(HaveLen(1))
				Expect(sph.requestedStopWaiting).To(BeTrue())
				Expect(mconn.written[0]).To(ContainSubstring("foobar1234567"))
			})

			It("sends a StreamFrame from a packet queued for retransmission", func() {
				f1 := frames.StreamFrame{
					StreamID: 0x5,
					Data:     []byte("foobar"),
				}
				f2 := frames.StreamFrame{
					StreamID: 0x7,
					Data:     []byte("loremipsum"),
				}
				p1 := ackhandler.Packet{
					PacketNumber:    0x1337,
					Frames:          []frames.Frame{&f1},
					EncryptionLevel: protocol.EncryptionForwardSecure,
				}
				p2 := ackhandler.Packet{
					PacketNumber:    0x1338,
					Frames:          []frames.Frame{&f2},
					EncryptionLevel: protocol.EncryptionForwardSecure,
				}
				sph.retransmissionQueue = []*ackhandler.Packet{&p1, &p2}

				err := sess.sendPacket()
				Expect(err).NotTo(HaveOccurred())
				Expect(mconn.written).To(HaveLen(1))
				Expect(mconn.written[0]).To(ContainSubstring("foobar"))
				Expect(mconn.written[0]).To(ContainSubstring("loremipsum"))
			})

			It("always attaches a StopWaiting to a packet that contains a retransmission", func() {
				f := &frames.StreamFrame{
					StreamID: 0x5,
					Data:     bytes.Repeat([]byte{'f'}, int(1.5*float32(protocol.MaxPacketSize))),
				}
				sess.streamFramer.AddFrameForRetransmission(f)

				err := sess.sendPacket()
				Expect(err).NotTo(HaveOccurred())
				Expect(mconn.written).To(HaveLen(2))
				sentPackets := sph.sentPackets
				Expect(sentPackets).To(HaveLen(2))
				_, ok := sentPackets[0].Frames[0].(*frames.StopWaitingFrame)
				Expect(ok).To(BeTrue())
				_, ok = sentPackets[1].Frames[0].(*frames.StopWaitingFrame)
				Expect(ok).To(BeTrue())
			})

			It("retransmits a WindowUpdate if it hasn't already sent a WindowUpdate with a higher ByteOffset", func() {
				_, err := sess.GetOrOpenStream(5)
				Expect(err).ToNot(HaveOccurred())
				fcm := mocks_fc.NewMockFlowControlManager(mockCtrl)
				sess.flowControlManager = fcm
				fcm.EXPECT().GetWindowUpdates()
				fcm.EXPECT().GetReceiveWindow(protocol.StreamID(5)).Return(protocol.ByteCount(0x1000), nil)
				wuf := &frames.WindowUpdateFrame{
					StreamID:   5,
					ByteOffset: 0x1000,
				}
				sph.retransmissionQueue = []*ackhandler.Packet{{
					Frames:          []frames.Frame{wuf},
					EncryptionLevel: protocol.EncryptionForwardSecure,
				}}
				err = sess.sendPacket()
				Expect(err).ToNot(HaveOccurred())
				Expect(sph.sentPackets).To(HaveLen(1))
				Expect(sph.sentPackets[0].Frames).To(ContainElement(wuf))
			})

			It("doesn't retransmit WindowUpdates if it already sent a WindowUpdate with a higher ByteOffset", func() {
				_, err := sess.GetOrOpenStream(5)
				Expect(err).ToNot(HaveOccurred())
				fcm := mocks_fc.NewMockFlowControlManager(mockCtrl)
				sess.flowControlManager = fcm
				fcm.EXPECT().GetWindowUpdates()
				fcm.EXPECT().GetReceiveWindow(protocol.StreamID(5)).Return(protocol.ByteCount(0x2000), nil)
				sph.retransmissionQueue = []*ackhandler.Packet{{
					Frames: []frames.Frame{&frames.WindowUpdateFrame{
						StreamID:   5,
						ByteOffset: 0x1000,
					}},
					EncryptionLevel: protocol.EncryptionForwardSecure,
				}}
				err = sess.sendPacket()
				Expect(err).ToNot(HaveOccurred())
				Expect(sph.sentPackets).To(BeEmpty())
			})

			It("doesn't retransmit WindowUpdates for closed streams", func() {
				str, err := sess.GetOrOpenStream(5)
				Expect(err).ToNot(HaveOccurred())
				// close the stream
				str.(*stream).sentFin()
				str.Close()
				str.(*stream).RegisterRemoteError(nil)
				sess.garbageCollectStreams()
				_, err = sess.flowControlManager.SendWindowSize(5)
				Expect(err).To(MatchError("Error accessing the flowController map."))
				sph.retransmissionQueue = []*ackhandler.Packet{{
					Frames: []frames.Frame{&frames.WindowUpdateFrame{
						StreamID:   5,
						ByteOffset: 0x1337,
					}},
					EncryptionLevel: protocol.EncryptionForwardSecure,
				}}
				err = sess.sendPacket()
				Expect(err).ToNot(HaveOccurred())
				Expect(sph.sentPackets).To(BeEmpty())
			})
		})
	})

	It("retransmits RTO packets", func() {
		n := protocol.PacketNumber(10)
		sess.packer.cryptoSetup = &mockCryptoSetup{encLevelSeal: protocol.EncryptionForwardSecure}
		// We simulate consistently low RTTs, so that the test works faster
		rtt := time.Millisecond
		sess.rttStats.UpdateRTT(rtt, 0, time.Now())
		Expect(sess.rttStats.SmoothedRTT()).To(Equal(rtt)) // make sure it worked
		sess.packer.packetNumberGenerator.next = n + 1
		// Now, we send a single packet, and expect that it was retransmitted later
		err := sess.sentPacketHandler.SentPacket(&ackhandler.Packet{
			PacketNumber: n,
			Length:       1,
			Frames: []frames.Frame{&frames.StreamFrame{
				Data: []byte("foobar"),
			}},
			EncryptionLevel: protocol.EncryptionForwardSecure,
		})
		Expect(err).NotTo(HaveOccurred())
		go sess.run()
		defer sess.Close(nil)
		sess.scheduleSending()
		Eventually(func() int { return len(mconn.written) }).ShouldNot(BeZero())
		Expect(mconn.written[0]).To(ContainSubstring("foobar"))
	})

	Context("scheduling sending", func() {
		BeforeEach(func() {
			sess.packer.cryptoSetup = &mockCryptoSetup{encLevelSeal: protocol.EncryptionForwardSecure}
		})

		It("sends after writing to a stream", func(done Done) {
			Expect(sess.sendingScheduled).NotTo(Receive())
			s, err := sess.GetOrOpenStream(3)
			Expect(err).NotTo(HaveOccurred())
			go func() {
				s.Write([]byte("foobar"))
				close(done)
			}()
			Eventually(sess.sendingScheduled).Should(Receive())
			s.(*stream).getDataForWriting(1000) // unblock
		})

		It("sets the timer to the ack timer", func() {
			rph := &mockReceivedPacketHandler{ackAlarm: time.Now().Add(10 * time.Millisecond)}
			rph.nextAckFrame = &frames.AckFrame{LargestAcked: 0x1337}
			sess.receivedPacketHandler = rph
			go sess.run()
			defer sess.Close(nil)
			time.Sleep(10 * time.Millisecond)
			Eventually(func() int { return len(mconn.written) }).ShouldNot(BeZero())
			Expect(mconn.written[0]).To(ContainSubstring(string([]byte{0x37, 0x13})))
		})

		Context("bundling of small packets", func() {
			It("bundles two small frames of different streams into one packet", func() {
				s1, err := sess.GetOrOpenStream(5)
				Expect(err).NotTo(HaveOccurred())
				s2, err := sess.GetOrOpenStream(7)
				Expect(err).NotTo(HaveOccurred())

				// Put data directly into the streams
				s1.(*stream).dataForWriting = []byte("foobar1")
				s2.(*stream).dataForWriting = []byte("foobar2")

				sess.scheduleSending()
				go sess.run()
				defer sess.Close(nil)

				Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(1))
				Expect(mconn.written[0]).To(ContainSubstring("foobar1"))
				Expect(mconn.written[0]).To(ContainSubstring("foobar2"))
			})

			It("sends out two big frames in two packets", func() {
				s1, err := sess.GetOrOpenStream(5)
				Expect(err).NotTo(HaveOccurred())
				s2, err := sess.GetOrOpenStream(7)
				Expect(err).NotTo(HaveOccurred())
				go sess.run()
				defer sess.Close(nil)
				go func() {
					defer GinkgoRecover()
					s1.Write(bytes.Repeat([]byte{'e'}, 1000))
				}()
				_, err = s2.Write(bytes.Repeat([]byte{'e'}, 1000))
				Expect(err).ToNot(HaveOccurred())
				Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(2))
			})

			It("sends out two small frames that are written to long after one another into two packets", func() {
				s, err := sess.GetOrOpenStream(5)
				Expect(err).NotTo(HaveOccurred())
				go sess.run()
				defer sess.Close(nil)
				_, err = s.Write([]byte("foobar1"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(1))
				_, err = s.Write([]byte("foobar2"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(2))
			})

			It("sends a queued ACK frame only once", func() {
				packetNumber := protocol.PacketNumber(0x1337)
				sess.receivedPacketHandler.ReceivedPacket(packetNumber, true)

				s, err := sess.GetOrOpenStream(5)
				Expect(err).NotTo(HaveOccurred())
				go sess.run()
				defer sess.Close(nil)
				_, err = s.Write([]byte("foobar1"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(1))
				_, err = s.Write([]byte("foobar2"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(2))
				Expect(mconn.written[0]).To(ContainSubstring(string([]byte{0x37, 0x13})))
				Expect(mconn.written[1]).ToNot(ContainSubstring(string([]byte{0x37, 0x13})))
			})
		})
	})

	It("closes when crypto stream errors", func() {
		testErr := errors.New("crypto setup error")
		cryptoSetup.handleErr = testErr
		var runErr error
		go func() {
			runErr = sess.run()
		}()
		Eventually(func() error { return runErr }).Should(HaveOccurred())
		Expect(runErr).To(MatchError(testErr))
	})

	Context("sending a Public Reset when receiving undecryptable packets during the handshake", func() {
		// sends protocol.MaxUndecryptablePackets+1 undecrytable packets
		// this completely fills up the undecryptable packets queue and triggers the public reset timer
		sendUndecryptablePackets := func() {
			for i := 0; i < protocol.MaxUndecryptablePackets+1; i++ {
				hdr := &PublicHeader{
					PacketNumber: protocol.PacketNumber(i + 1),
				}
				sess.handlePacket(&receivedPacket{
					publicHeader: hdr,
					remoteAddr:   &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1234},
					data:         []byte("foobar"),
				})
			}
		}

		BeforeEach(func() {
			sess.unpacker = &mockUnpacker{unpackErr: qerr.Error(qerr.DecryptionFailure, "")}
			sess.cryptoSetup = &mockCryptoSetup{}
		})

		It("doesn't immediately send a Public Reset after receiving too many undecryptable packets", func() {
			go sess.run()
			sendUndecryptablePackets()
			sess.scheduleSending()
			Consistently(func() [][]byte { return mconn.written }).Should(HaveLen(0))
		})

		It("sets a deadline to send a Public Reset after receiving too many undecryptable packets", func() {
			go sess.run()
			sendUndecryptablePackets()
			Eventually(func() time.Time { return sess.receivedTooManyUndecrytablePacketsTime }).Should(BeTemporally("~", time.Now(), 20*time.Millisecond))
			sess.Close(nil)
		})

		It("drops undecryptable packets when the undecrytable packet queue is full", func() {
			go sess.run()
			sendUndecryptablePackets()
			Eventually(func() []*receivedPacket { return sess.undecryptablePackets }).Should(HaveLen(protocol.MaxUndecryptablePackets))
			// check that old packets are kept, and the new packets are dropped
			Expect(sess.undecryptablePackets[0].publicHeader.PacketNumber).To(Equal(protocol.PacketNumber(1)))
			sess.Close(nil)
		})

		It("sends a Public Reset after a timeout", func() {
			go sess.run()
			sendUndecryptablePackets()
			Eventually(func() time.Time { return sess.receivedTooManyUndecrytablePacketsTime }).Should(BeTemporally("~", time.Now(), 10*time.Millisecond))
			// speed up this test by manually setting back the time when too many packets were received
			sess.receivedTooManyUndecrytablePacketsTime = time.Now().Add(-protocol.PublicResetTimeout)
			time.Sleep(10 * time.Millisecond) // wait for the run loop to spin up
			sess.scheduleSending()            // wake up the run loop
			Eventually(func() [][]byte { return mconn.written }).Should(HaveLen(1))
			Expect(mconn.written[0]).To(ContainSubstring("PRST"))
			Eventually(sess.runClosed).Should(BeClosed())
		})

		It("doesn't send a Public Reset if decrypting them suceeded during the timeout", func() {
			go sess.run()
			sess.receivedTooManyUndecrytablePacketsTime = time.Now().Add(-protocol.PublicResetTimeout).Add(-time.Millisecond)
			sess.scheduleSending() // wake up the run loop
			// there are no packets in the undecryptable packet queue
			// in reality, this happens when the trial decryption succeeded during the Public Reset timeout
			Consistently(func() [][]byte { return mconn.written }).ShouldNot(HaveLen(1))
			Expect(sess.runClosed).ToNot(Receive())
			sess.Close(nil)
		})

		It("ignores undecryptable packets after the handshake is complete", func() {
			sess.handshakeComplete = true
			go sess.run()
			sendUndecryptablePackets()
			Consistently(sess.undecryptablePackets).Should(BeEmpty())
			Expect(sess.Close(nil)).To(Succeed())
		})

		It("unqueues undecryptable packets for later decryption", func() {
			sess.undecryptablePackets = []*receivedPacket{{
				publicHeader: &PublicHeader{PacketNumber: protocol.PacketNumber(42)},
			}}
			Expect(sess.receivedPackets).NotTo(Receive())
			sess.tryDecryptingQueuedPackets()
			Expect(sess.undecryptablePackets).To(BeEmpty())
			Expect(sess.receivedPackets).To(Receive())
		})
	})

	It("send a handshake event on the handshakeChan when the AEAD changes to secure", func(done Done) {
		go sess.run()
		aeadChanged <- protocol.EncryptionSecure
		Eventually(handshakeChan).Should(Receive(&handshakeEvent{encLevel: protocol.EncryptionSecure}))
		Expect(sess.Close(nil)).To(Succeed())
		close(done)
	})

	It("send a handshake event on the handshakeChan when the AEAD changes to forward-secure", func(done Done) {
		go sess.run()
		aeadChanged <- protocol.EncryptionForwardSecure
		Eventually(handshakeChan).Should(Receive(&handshakeEvent{encLevel: protocol.EncryptionForwardSecure}))
		Expect(sess.Close(nil)).To(Succeed())
		close(done)
	})

	It("closes the handshakeChan when the handshake completes", func(done Done) {
		go sess.run()
		close(aeadChanged)
		Eventually(handshakeChan).Should(BeClosed())
		Expect(sess.Close(nil)).To(Succeed())
		close(done)
	})

	It("passes errors to the handshakeChan", func(done Done) {
		testErr := errors.New("handshake error")
		go sess.run()
		Expect(sess.Close(nil)).To(Succeed())
		Expect(handshakeChan).To(Receive(&handshakeEvent{err: testErr}))
		close(done)
	})

	It("does not block if an error occurs", func(done Done) {
		// this test basically tests that the handshakeChan has a capacity of 3
		// The session needs to run (and close) properly, even if no one is receiving from the handshakeChan
		go sess.run()
		aeadChanged <- protocol.EncryptionSecure
		aeadChanged <- protocol.EncryptionForwardSecure
		Expect(sess.Close(nil)).To(Succeed())
		close(done)
	})

	Context("keep-alives", func() {
		It("sends a PING", func() {
			sess.handshakeComplete = true
			sess.config.KeepAlive = true
			sess.lastNetworkActivityTime = time.Now().Add(-(sess.idleTimeout() / 2))
			go sess.run()
			defer sess.Close(nil)
			time.Sleep(60 * time.Millisecond)
			Eventually(func() [][]byte { return mconn.written }).ShouldNot(BeEmpty())
			Eventually(func() byte {
				// -12 because of the crypto tag. This should be 7 (the frame id for a ping frame).
				s := mconn.written[0]
				return s[len(s)-12-1]
			}).Should(Equal(byte(0x07)))
		})

		It("doesn't send a PING packet if keep-alive is disabled", func() {
			sess.handshakeComplete = true
			sess.lastNetworkActivityTime = time.Now().Add(-(sess.idleTimeout() / 2))
			go sess.run()
			defer sess.Close(nil)
			Consistently(func() [][]byte { return mconn.written }).Should(BeEmpty())
		})

		It("doesn't send a PING if the handshake isn't completed yet", func() {
			sess.handshakeComplete = false
			sess.config.KeepAlive = true
			sess.lastNetworkActivityTime = time.Now().Add(-(sess.idleTimeout() / 2))
			go sess.run()
			defer sess.Close(nil)
			Consistently(func() [][]byte { return mconn.written }).Should(BeEmpty())
		})
	})

	Context("timeouts", func() {
		It("times out due to no network activity", func(done Done) {
			sess.lastNetworkActivityTime = time.Now().Add(-time.Hour)
			err := sess.run() // Would normally not return
			Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.NetworkIdleTimeout))
			Expect(mconn.written[0]).To(ContainSubstring("No recent network activity."))
			Expect(sess.runClosed).To(BeClosed())
			close(done)
		})

		It("times out due to non-completed crypto handshake", func(done Done) {
			sess.sessionCreationTime = time.Now().Add(-protocol.DefaultHandshakeTimeout).Add(-time.Second)
			err := sess.run() // Would normally not return
			Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.HandshakeTimeout))
			Expect(mconn.written[0]).To(ContainSubstring("Crypto handshake did not complete in time."))
			Expect(sess.runClosed).To(BeClosed())
			close(done)
		})

		It("does not use ICSL before handshake", func(done Done) {
			sess.lastNetworkActivityTime = time.Now().Add(-time.Minute)
			mockCpm = mocks.NewMockConnectionParametersManager(mockCtrl)
			mockCpm.EXPECT().GetIdleConnectionStateLifetime().Return(9999 * time.Second).AnyTimes()
			mockCpm.EXPECT().TruncateConnectionID().Return(false).AnyTimes()
			sess.connectionParameters = mockCpm
			sess.packer.connectionParameters = mockCpm
			err := sess.run() // Would normally not return
			Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.NetworkIdleTimeout))
			Expect(mconn.written[0]).To(ContainSubstring("No recent network activity."))
			Expect(sess.runClosed).To(BeClosed())
			close(done)
		})

		It("uses ICSL after handshake", func(done Done) {
			close(aeadChanged)
			mockCpm = mocks.NewMockConnectionParametersManager(mockCtrl)
			mockCpm.EXPECT().GetIdleConnectionStateLifetime().Return(0 * time.Second)
			mockCpm.EXPECT().TruncateConnectionID().Return(false).AnyTimes()
			sess.connectionParameters = mockCpm
			sess.packer.connectionParameters = mockCpm
			mockCpm.EXPECT().GetIdleConnectionStateLifetime().Return(0 * time.Second).AnyTimes()
			err := sess.run() // Would normally not return
			Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.NetworkIdleTimeout))
			Expect(mconn.written[0]).To(ContainSubstring("No recent network activity."))
			Expect(sess.runClosed).To(BeClosed())
			close(done)
		})
	})

	It("stores up to MaxSessionUnprocessedPackets packets", func(done Done) {
		// Nothing here should block
		for i := protocol.PacketNumber(0); i < protocol.MaxSessionUnprocessedPackets+10; i++ {
			sess.handlePacket(&receivedPacket{})
		}
		close(done)
	}, 0.5)

	Context("getting streams", func() {
		It("returns a new stream", func() {
			str, err := sess.GetOrOpenStream(11)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			Expect(str.StreamID()).To(Equal(protocol.StreamID(11)))
		})

		It("returns a nil-value (not an interface with value nil) for closed streams", func() {
			_, err := sess.GetOrOpenStream(9)
			Expect(err).ToNot(HaveOccurred())
			sess.streamsMap.RemoveStream(9)
			sess.garbageCollectStreams()
			Expect(sess.streamsMap.GetOrOpenStream(9)).To(BeNil())
			str, err := sess.GetOrOpenStream(9)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(BeNil())
			// make sure that the returned value is a plain nil, not an Stream with value nil
			_, ok := str.(Stream)
			Expect(ok).To(BeFalse())
		})

		// all relevant tests for this are in the streamsMap
		It("opens streams synchronously", func() {
			str, err := sess.OpenStreamSync()
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("counting streams", func() {
		It("errors when too many streams are opened", func() {
			for i := 0; i < 110; i++ {
				_, err := sess.GetOrOpenStream(protocol.StreamID(i*2 + 1))
				Expect(err).NotTo(HaveOccurred())
			}
			_, err := sess.GetOrOpenStream(protocol.StreamID(301))
			Expect(err).To(MatchError(qerr.TooManyOpenStreams))
		})

		It("does not error when many streams are opened and closed", func() {
			for i := 2; i <= 1000; i++ {
				s, err := sess.GetOrOpenStream(protocol.StreamID(i*2 + 1))
				Expect(err).NotTo(HaveOccurred())
				err = s.Close()
				Expect(err).NotTo(HaveOccurred())
				s.(*stream).sentFin()
				s.(*stream).CloseRemote(0)
				_, err = s.Read([]byte("a"))
				Expect(err).To(MatchError(io.EOF))
				sess.garbageCollectStreams()
			}
		})
	})

	Context("ignoring errors", func() {
		It("ignores duplicate acks", func() {
			sess.sentPacketHandler.SentPacket(&ackhandler.Packet{
				PacketNumber: 1,
				Length:       1,
			})
			err := sess.handleFrames([]frames.Frame{&frames.AckFrame{
				LargestAcked: 1,
			}})
			Expect(err).NotTo(HaveOccurred())
			err = sess.handleFrames([]frames.Frame{&frames.AckFrame{
				LargestAcked: 1,
			}})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("window updates", func() {
		It("gets stream level window updates", func() {
			err := sess.flowControlManager.AddBytesRead(1, protocol.ReceiveStreamFlowControlWindow)
			Expect(err).NotTo(HaveOccurred())
			frames := sess.getWindowUpdateFrames()
			Expect(frames).To(HaveLen(1))
			Expect(frames[0].StreamID).To(Equal(protocol.StreamID(1)))
			Expect(frames[0].ByteOffset).To(Equal(protocol.ReceiveStreamFlowControlWindow * 2))
		})

		It("gets connection level window updates", func() {
			_, err := sess.GetOrOpenStream(5)
			Expect(err).NotTo(HaveOccurred())
			err = sess.flowControlManager.AddBytesRead(5, protocol.ReceiveConnectionFlowControlWindow)
			Expect(err).NotTo(HaveOccurred())
			frames := sess.getWindowUpdateFrames()
			Expect(frames).To(HaveLen(1))
			Expect(frames[0].StreamID).To(Equal(protocol.StreamID(0)))
			Expect(frames[0].ByteOffset).To(Equal(protocol.ReceiveConnectionFlowControlWindow * 2))
		})
	})

	It("returns the local address", func() {
		addr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1337}
		mconn.localAddr = addr
		Expect(sess.LocalAddr()).To(Equal(addr))
	})

	It("returns the remote address", func() {
		addr := &net.UDPAddr{IP: net.IPv4(1, 2, 7, 1), Port: 7331}
		mconn.remoteAddr = addr
		Expect(sess.RemoteAddr()).To(Equal(addr))
	})
})

var _ = Describe("Client Session", func() {
	var (
		sess        *session
		mconn       *mockConnection
		aeadChanged chan<- protocol.EncryptionLevel

		cryptoSetup *mockCryptoSetup
	)

	BeforeEach(func() {
		Eventually(areSessionsRunning).Should(BeFalse())

		cryptoSetup = &mockCryptoSetup{}
		newCryptoSetupClient = func(
			_ string,
			_ protocol.ConnectionID,
			_ protocol.VersionNumber,
			_ io.ReadWriter,
			_ *tls.Config,
			_ handshake.ConnectionParametersManager,
			aeadChangedP chan<- protocol.EncryptionLevel,
			_ *handshake.TransportParameters,
			_ []protocol.VersionNumber,
		) (handshake.CryptoSetup, error) {
			aeadChanged = aeadChangedP
			return cryptoSetup, nil
		}

		mconn = &mockConnection{
			remoteAddr: &net.UDPAddr{},
		}
		sessP, _, err := newClientSession(
			mconn,
			"hostname",
			protocol.Version35,
			0,
			nil,
			populateClientConfig(&Config{}),
			nil,
		)
		sess = sessP.(*session)
		Expect(err).ToNot(HaveOccurred())
		Expect(sess.streamsMap.openStreams).To(HaveLen(1)) // Crypto stream
	})

	AfterEach(func() {
		newCryptoSetupClient = handshake.NewCryptoSetupClient
	})

	Context("receiving packets", func() {
		var hdr *PublicHeader

		BeforeEach(func() {
			hdr = &PublicHeader{PacketNumberLen: protocol.PacketNumberLen6}
			sess.unpacker = &mockUnpacker{}
		})

		It("passes the diversification nonce to the cryptoSetup", func() {
			go sess.run()
			hdr.PacketNumber = 5
			hdr.DiversificationNonce = []byte("foobar")
			err := sess.handlePacketImpl(&receivedPacket{publicHeader: hdr})
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() []byte { return cryptoSetup.divNonce }).Should(Equal(hdr.DiversificationNonce))
			Expect(sess.Close(nil)).To(Succeed())
		})
	})

	It("does not block if an error occurs", func(done Done) {
		// this test basically tests that the handshakeChan has a capacity of 3
		// The session needs to run (and close) properly, even if no one is receiving from the handshakeChan
		go sess.run()
		aeadChanged <- protocol.EncryptionSecure
		aeadChanged <- protocol.EncryptionForwardSecure
		Expect(sess.Close(nil)).To(Succeed())
		close(done)
	})
})
