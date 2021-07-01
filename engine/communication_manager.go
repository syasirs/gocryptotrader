package engine

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/thrasher-corp/gocryptotrader/communications"
	"github.com/thrasher-corp/gocryptotrader/communications/base"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// CommunicationsManagerName is an exported subsystem name
const CommunicationsManagerName = "communications"

// CommunicationManager ensures operations of communications
type CommunicationManager struct {
	started  int32
	shutdown chan struct{}
	relayMsg chan base.Event
	comms    *communications.Communications
}

var errNilConfig = errors.New("received nil communications config")

// SetupCommunicationManager creates a communications manager
func SetupCommunicationManager(cfg *base.CommunicationsConfig) (*CommunicationManager, error) {
	if cfg == nil {
		return nil, errNilConfig
	}
	manager := &CommunicationManager{
		shutdown: make(chan struct{}),
		relayMsg: make(chan base.Event),
	}
	var err error
	manager.comms, err = communications.NewComm(cfg)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

// IsRunning safely checks whether the subsystem is running
func (m *CommunicationManager) IsRunning() bool {
	if m == nil {
		return false
	}
	return atomic.LoadInt32(&m.started) == 1
}

// Start runs the subsystem
func (m *CommunicationManager) Start() error {
	if m == nil {
		return fmt.Errorf("communications manager server %w", ErrNilSubsystem)
	}
	if !atomic.CompareAndSwapInt32(&m.started, 0, 1) {
		return fmt.Errorf("communications manager %w", ErrSubSystemAlreadyStarted)
	}
	log.CommunicationMgr.Debugf("Communications manager %s", MsgSubSystemStarting)
	m.shutdown = make(chan struct{})
	go m.run()
	return nil
}

// GetStatus returns the status of communications
func (m *CommunicationManager) GetStatus() (map[string]base.CommsStatus, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("communications manager %w", ErrSubSystemNotStarted)
	}
	return m.comms.GetStatus(), nil
}

// Stop attempts to shutdown the subsystem
func (m *CommunicationManager) Stop() error {
	if m == nil {
		return fmt.Errorf("communications manager server %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return fmt.Errorf("communications manager %w", ErrSubSystemNotStarted)
	}
	defer func() {
		atomic.CompareAndSwapInt32(&m.started, 1, 0)
	}()
	close(m.shutdown)
	log.CommunicationMgr.Debugf("Communications manager %s", MsgSubSystemShuttingDown)
	return nil
}

// PushEvent pushes an event to the communications relay
func (m *CommunicationManager) PushEvent(evt base.Event) {
	if !m.IsRunning() {
		return
	}
	select {
	case m.relayMsg <- evt:
	default:
		log.CommunicationMgr.Errorf("Failed to send, no receiver when pushing event [%v]", evt)
	}
}

// run takes awaiting messages and pushes them to be handled by communications
func (m *CommunicationManager) run() {
	log.Global.Debugf("Communications manager %s", MsgSubSystemStarted)
	defer func() {
		// TO-DO shutdown comms connections for connected services (Slack etc)
		log.CommunicationMgr.Debugf("Communications manager %s", MsgSubSystemShutdown)
	}()

	for {
		select {
		case msg := <-m.relayMsg:
			m.comms.PushEvent(msg)
		case <-m.shutdown:
			return
		}
	}
}
