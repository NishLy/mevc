package rtc

import (
	"sort"
	"sync"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
)

var MAX_STREAMS_PER_PAGE = 4

type PaginationOrdering struct {
	index  int
	tracks []*TrackRouter
}

type SessionParticipantData struct {
	ClientID   string `json:"clientId"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	Muted      bool   `json:"muted"`
	VideoOff   bool   `json:"videoOff"`
	RaisedHand bool   `json:"raisedHand"`
}

type SessionManager interface {
	GetGroupId() string

	AddSession(session Session, wsID string) bool
	GetSession(clientId string) (Session, bool)
	GetSessions() []Session

	CloseAll() error

	GetClientIDFromWsID(wsID string) (string, bool)
	GetSessionByWsID(wsID string) (Session, bool)

	AddRouter(trackId string, router *TrackRouter)
	GetRouter(streamID string) (*TrackRouter, bool)
	RemoveRouter(streamID string)

	MoveToRoom(session Session)
	KickFromLoby(session Session)
	GetLobbySessions() []LobySession

	// AddTemporaryMetadata(streamID string, metadata SessionTrackMetadata)
	// GetTemporaryMetadata(streamID string) (SessionTrackMetadata, bool)
	// RemoveTemporaryMetadata(streamID string)

	RemoveFromSessionManager(clientID string)

	GetRouterPaginated(excludedCLientId *string, page int, pageSize int) (map[string][]*TrackRouter, int)
	SuscribeToPageinatedRouters(session Session, page int, pageSize int) error
	HighlightStreamRouters(streamIDS []string, priority int)

	AddClientGroupedStreams(clientID string, trackId string, router *TrackRouter)
	RemoveClientSubscribedTrack(clientID string, trackId string)
	GetClientGroupedStreams(clientID string) map[string][]*TrackRouter

	GetGroupedRouters() map[string][]*TrackRouter
	SetEmitFCN(func(event string, args ...interface{}))

	GetParticipantsData() []SessionParticipantData

	GetChatService() *ChatService
}

type LobySession struct {
	wsID    string
	session Session
}

type RoomData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	HostID      string `json:"host_id"`
	IsPrivate   bool   `json:"is_private"`
}

type RoomCurrentState struct {
	MaxiumPerPage              int `json:"maxium_per_page"`
	CurrentTotalParticipants   int `json:"current_total_participants"`
	CurrentTotalGroupedStreams int `json:"current_total_grouped_streams"`
}

type sessionManager struct {
	// own data
	id         string
	autoClose  bool
	autoAccept bool
	// session management
	lobby                   map[string]LobySession
	sessions                map[string]Session
	wsToClientId            map[string]string
	sessionSubscribedTracks map[string]map[string][]*TrackRouter // clientID -> streamGroupID  -> router
	// track management
	routers              []*TrackRouter
	hasChangedRouters    bool
	cachedOrderedRouters []string
	groupedRouters       map[string][]*TrackRouter
	// utility
	mu      sync.RWMutex
	EmitFCN func(event string, args ...interface{})
	// chat management
	chatService *ChatService

	// clean up management
	selfDeleteTimer *time.Timer
}

func NewSessionManager(id string, autoAccept bool) SessionManager {
	return &sessionManager{
		id:                      id,
		sessions:                make(map[string]Session),
		wsToClientId:            make(map[string]string),
		routers:                 make([]*TrackRouter, 0),
		autoClose:               true,
		autoAccept:              autoAccept,
		lobby:                   make(map[string]LobySession),
		mu:                      sync.RWMutex{},
		sessionSubscribedTracks: make(map[string]map[string][]*TrackRouter),
		groupedRouters:          make(map[string][]*TrackRouter),
		hasChangedRouters:       true,
		cachedOrderedRouters:    make([]string, 0),
		chatService:             NewChatService(),
	}
}

func (sm *sessionManager) GetChatService() *ChatService {
	return sm.chatService
}

func (sm *sessionManager) GetParticipantsData() []SessionParticipantData {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	participants := make([]SessionParticipantData, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessionState := session.GetCurrentState()
		participants = append(participants, SessionParticipantData{
			ClientID:   session.GetClientId(),
			Username:   session.GetUsername(),
			Role:       "speaker", // Not implemented yet, defaulting to speaker for now, can be extended in the future to support different roles and permissions
			Muted:      sessionState.IsMuted,
			VideoOff:   sessionState.IsVideoOff,
			RaisedHand: sessionState.IsRaisedHand,
		})
	}

	return participants
}

func (sm *sessionManager) SetEmitFCN(fcn func(event string, args ...interface{})) {
	sm.EmitFCN = fcn
}

func (sm *sessionManager) GetGroupedRouters() map[string][]*TrackRouter {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.groupedRouters
}

func (sm *sessionManager) GetClientGroupedStreams(clientID string) map[string][]*TrackRouter {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	tracks, exist := sm.sessionSubscribedTracks[clientID]
	if !exist {
		return make(map[string][]*TrackRouter)
	}
	return tracks
}

func (sm *sessionManager) RemoveClientSubscribedTrack(clientID string, trackId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exist := sm.sessionSubscribedTracks[clientID]; exist {
		delete(sm.sessionSubscribedTracks[clientID], trackId)
	}
}

func (sm *sessionManager) AddClientGroupedStreams(clientID string, trackId string, router *TrackRouter) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, exist := sm.sessionSubscribedTracks[clientID]

	if !exist {
		sm.sessionSubscribedTracks[clientID] = make(map[string][]*TrackRouter)
	}

	if _, exist := sm.sessionSubscribedTracks[clientID]; !exist {
		sm.sessionSubscribedTracks[clientID] = make(map[string][]*TrackRouter)
	}

	sm.sessionSubscribedTracks[clientID][trackId] = append(sm.sessionSubscribedTracks[clientID][trackId], router)
}

func (sm *sessionManager) GetLobbySessions() []LobySession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	lobbySessions := make([]LobySession, 0, len(sm.lobby))

	for _, lobbySession := range sm.lobby {
		lobbySessions = append(lobbySessions, lobbySession)
	}
	return lobbySessions
}

func (sm *sessionManager) MoveToRoom(session Session) {
	sm.mu.Lock()
	lobbySession, exist := sm.lobby[session.GetClientId()]
	if !exist {
		logger.Sugar.Warnf("No lobby session found for client %s, rejecting session", session.GetClientId())
	}

	delete(sm.lobby, session.GetClientId())
	sm.mu.Unlock()
	sm.AddSession(lobbySession.session, lobbySession.wsID)
}

func (sm *sessionManager) KickFromLoby(session Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, exist := sm.lobby[session.GetClientId()]
	if !exist {
		return
	}
	delete(sm.lobby, session.GetClientId())
	session.Close()
}

func (sm *sessionManager) GetGroupId() string {
	return sm.id
}

func (sm *sessionManager) GetSessions() []Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *sessionManager) GetSession(clientId string) (Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[clientId]
	return session, exists
}

func (sm *sessionManager) AddSession(session Session, wsID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.wsToClientId[wsID] = session.GetClientId()
	if !sm.autoAccept {
		sm.lobby[session.GetClientId()] = LobySession{
			wsID:    wsID,
			session: session,
		}
		logger.Sugar.Infof("Added session for client %s to lobby", session.GetClientId())
		return false
	}

	sm.sessions[session.GetClientId()] = session

	sm.EmitFCN("room_state_changed", RoomCurrentState{
		MaxiumPerPage:              MAX_STREAMS_PER_PAGE,
		CurrentTotalParticipants:   len(sm.sessions),
		CurrentTotalGroupedStreams: len(sm.groupedRouters),
	})

	return true
}

func (sm *sessionManager) CloseAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, session := range sm.sessions {
		session.Close()
	}
	sm.sessions = make(map[string]Session)
	sm.wsToClientId = make(map[string]string)
	for _, router := range sm.routers {
		router.Close()
	}
	sm.routers = make([]*TrackRouter, 0)
	sm.groupedRouters = make(map[string][]*TrackRouter)
	sm.chatService.Close()
	return nil
}

func (sm *sessionManager) GetClientIDFromWsID(wsID string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	clientID, exists := sm.wsToClientId[wsID]
	return clientID, exists
}

func (sm *sessionManager) GetSessionByWsID(wsID string) (Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	clientID, exists := sm.wsToClientId[wsID]
	if !exists {
		return nil, false
	}

	lobbySession, existInLobby := sm.lobby[clientID]
	_, existInSessions := sm.sessions[clientID]

	if existInLobby {
		return lobbySession.session, true
	}
	if !existInSessions {
		return nil, false
	}

	session, exists := sm.sessions[clientID]
	return session, exists
}

func (sm *sessionManager) AddRouter(trackId string, router *TrackRouter) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var isExisting bool
	for _, r := range sm.routers {
		if r.streamID == trackId {
			isExisting = true
			break
		}
	}

	if !isExisting {
		sm.routers = append(sm.routers, router)
		sm.hasChangedRouters = true
	}

	sm.EmitFCN("room_state_changed", RoomCurrentState{
		MaxiumPerPage:              MAX_STREAMS_PER_PAGE,
		CurrentTotalParticipants:   len(sm.sessions),
		CurrentTotalGroupedStreams: len(sm.groupedRouters),
	})
}

func (sm *sessionManager) GetRouter(streamId string) (*TrackRouter, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, router := range sm.routers {
		if router.streamID == streamId {
			return router, true
		}
	}
	return nil, false
}

func (sm *sessionManager) RemoveRouter(streamId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, router := range sm.routers {
		if router.streamID == streamId {
			router.Close()
			sm.routers = append(sm.routers[:i], sm.routers[i+1:]...)
			sm.hasChangedRouters = true
			break
		}
	}

	sm.EmitFCN("room_state_changed", RoomCurrentState{
		MaxiumPerPage:              MAX_STREAMS_PER_PAGE,
		CurrentTotalParticipants:   len(sm.sessions),
		CurrentTotalGroupedStreams: len(sm.groupedRouters),
	})
}

func (sm *sessionManager) RemoveFromSessionManager(clientID string) {
	session, exists := sm.GetSession(clientID)
	if !exists || session == nil {
		return
	}

	var routersToClose []struct {
		streamGroupId string
		streamID      string
	}

	sm.mu.RLock()
	for _, router := range sm.routers {
		if router.publisherID == session.GetClientId() {
			routersToClose = append(routersToClose, struct {
				streamGroupId string
				streamID      string
			}{
				streamGroupId: router.streamGroupId,
				streamID:      router.streamID,
			})
		}

		// Remove the viewer from all routers
		if err := router.RemoveViewer(clientID); err != nil {
			logger.Sugar.Warnf("Failed to remove viewer %s from router for stream %s: %v", clientID, router.incomingTrack.StreamID(), err)
		}
	}
	sm.mu.RUnlock()

	delete(sm.sessionSubscribedTracks, clientID) // Remove the client's subscribed tracks management since the client is leaving

	for _, trackId := range routersToClose {
		sm.RemoveRouter(trackId.streamID)
	}

	delete(sm.sessions, clientID)

	for _, session := range sm.sessions {
		storage := sm.GetClientGroupedStreams(session.GetClientId())

		// Check if this session is subscribed to any of the removed routers
		hasRemovedTrack := false
		for _, router := range routersToClose {
			if _, exist := storage[router.streamGroupId]; exist {
				sm.RemoveClientSubscribedTrack(session.GetClientId(), router.streamGroupId) // Clean up first
				hasRemovedTrack = true
			}
		}

		// Only resubscribe once per session, after all removed tracks are cleaned up
		if hasRemovedTrack {
			sm.SuscribeToPageinatedRouters(session, session.GetCurrentViewPage(), MAX_STREAMS_PER_PAGE)
		}
	}

	for wsID, cid := range sm.wsToClientId {
		if cid == clientID {
			delete(sm.wsToClientId, wsID)
			break
		}
	}

	session.Close()

	if len(sm.GetSessions()) == 0 {
		if sm.selfDeleteTimer != nil {
			sm.selfDeleteTimer.Stop() // Stop any existing timer to avoid multiple timers running if multiple clients leave in quick succession
		}

		sm.selfDeleteTimer = time.AfterFunc(5*time.Minute, func() {
			if len(sm.GetSessions()) == 0 {
				logger.Sugar.Infof("No sessions rejoined within the grace period, performing final cleanup for session manager %s", sm.id)
				sm.CloseAll()                       // Ensure all resources are cleaned up
				delete(GlobalSessionManager, sm.id) // Remove the session manager from the global map to allow garbage collection
			} else {
				if sm.selfDeleteTimer != nil {
					sm.selfDeleteTimer.Stop()
				}
			}
		})
	}

	sm.EmitFCN("room_state_changed", RoomCurrentState{
		MaxiumPerPage:              MAX_STREAMS_PER_PAGE,
		CurrentTotalParticipants:   len(sm.sessions),
		CurrentTotalGroupedStreams: len(sm.groupedRouters),
	})

	logger.Sugar.Infof("Removed session for client %s", clientID)
}

func (sm *sessionManager) GetRouterPaginated(excludedClientId *string, page int, pageSize int) (map[string][]*TrackRouter, int) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.hasChangedRouters {
		tempGroups := make(map[string]PaginationOrdering)
		sm.groupedRouters = make(map[string][]*TrackRouter)

		for index, router := range sm.routers {
			existingGroup, exist := tempGroups[router.streamGroupId]
			if !exist {
				existingGroup = PaginationOrdering{
					index:  index,
					tracks: []*TrackRouter{router},
				}
			} else {
				existingGroup.tracks = append(existingGroup.tracks, router)
			}
			tempGroups[router.streamGroupId] = existingGroup
		}

		keys := make([]string, 0, len(tempGroups))
		for key, group := range tempGroups {
			keys = append(keys, key)
			sm.groupedRouters[key] = group.tracks
		}

		sort.Slice(keys, func(i, j int) bool {
			return tempGroups[keys[i]].index < tempGroups[keys[j]].index
		})

		sm.cachedOrderedRouters = keys
		sm.hasChangedRouters = false
	}

	// Filter and paginate at read time
	filteredKeys := sm.cachedOrderedRouters
	if excludedClientId != nil {
		filteredKeys = make([]string, 0, len(sm.cachedOrderedRouters))
		for _, k := range sm.cachedOrderedRouters {
			hasNonExcluded := false
			for _, router := range sm.groupedRouters[k] {
				if router.publisherID != *excludedClientId {
					hasNonExcluded = true
					break
				}
			}
			if hasNonExcluded {
				filteredKeys = append(filteredKeys, k)
			}
		}
	}

	total := len(filteredKeys)
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= total {
		return map[string][]*TrackRouter{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pagedKeys := filteredKeys[start:end]
	result := make(map[string][]*TrackRouter)
	for _, k := range pagedKeys {
		tracks := sm.groupedRouters[k]
		if excludedClientId != nil {
			filtered := make([]*TrackRouter, 0, len(tracks))
			for _, router := range tracks {
				if router.publisherID != *excludedClientId {
					filtered = append(filtered, router)
				}
			}
			result[k] = filtered
		} else {
			result[k] = tracks
		}
	}

	return result, total
}

func (sm *sessionManager) SuscribeToPageinatedRouters(session Session, page int, pageSize int) error {
	excludedClientId := session.GetClientId()
	groupedRouters, _ := sm.GetRouterPaginated(&excludedClientId, page, pageSize)

	if len(groupedRouters) == 0 {
		return nil
	}

	for key, routers := range sm.GetClientGroupedStreams(session.GetClientId()) {
		// If the client is already subscribed to a router that is in the new paginated set of routers, we keep the subscription and don't remove it to avoid unnecessary renegotiation and resource usage
		if _, ok := groupedRouters[key]; ok {
			continue
		}

		for _, router := range routers {
			router.RemoveViewer(session.GetClientId())                             // First remove the viewer from all currently subscribed routers to avoid duplicates
			sm.RemoveClientSubscribedTrack(session.GetClientId(), router.streamID) // Remove all currently subscribed tracks from the session's management since we're going to resubscribe to a new paginated set of tracks
			session.Emit("track_removed", router.streamGroupId)                    // Emit track_removed event for all currently subscribed tracks so the client can remove them from the UI
		}
	}

	logger.Sugar.Infof("Client %s subscribing to page %d with %d stream groups %v", session.GetClientId(), page, len(groupedRouters), groupedRouters)

	var shouldRenegotiate bool
	for _, routers := range groupedRouters {
		for _, router := range routers {
			if router.publisherID == session.GetClientId() {
				continue // Don't subscribe the publisher to their own tracks
			}

			err := router.AddViewer(session)

			if err != nil {
				continue
			}

			shouldRenegotiate = true
			// Add the router to the session's current subscribed tracks so it can manage it (e.g. remove it when unsubscribing or when the session is closed)
			sm.AddClientGroupedStreams(session.GetClientId(), router.streamGroupId, router)
			session.Emit("new_track", router.streamID, router.metadata) // Emit metadata for the track to the client so it can render it in the UI
		}
	}

	logger.Sugar.Infof("Client %s subscribed to page %d with %d stream groups regenerated %v", session.GetClientId(), page, len(groupedRouters), shouldRenegotiate)

	if shouldRenegotiate {
		if err := session.Renegotiate(nil); err != nil {
			logger.Sugar.Warnf("Failed to renegotiate session for client %s after subscribing to paginated routers: %v", session.GetClientId(), err)
			return err
		}
	}
	return nil
}

// priority 1 = host, 2 = speaker, 3 = all, 4 = none

func (sm *sessionManager) HighlightStreamRouters(streamIDS []string, priority int) {
	sm.mu.Lock()

	var highlightedRouters []*TrackRouter

	var hostPriorities []*TrackRouter
	var speakerPriorities []*TrackRouter
	var allPriorities []*TrackRouter
	var nonePriorities []*TrackRouter

	highlightMap := make(map[string]struct{})
	for _, id := range streamIDS {
		highlightMap[id] = struct{}{}
	}

	for _, router := range sm.routers {
		if _, shouldHighlight := highlightMap[router.streamID]; shouldHighlight {
			highlightedRouters = append(highlightedRouters, router)
			continue
		}

		switch router.priority {
		case 1:
			hostPriorities = append(hostPriorities, router)
		case 2:
			speakerPriorities = append(speakerPriorities, router)
		case 3:
			allPriorities = append(allPriorities, router)
		case 4:
			nonePriorities = append(nonePriorities, router)
		}
	}

	sm.routers = sm.routers[:0] // Clear the slice while keeping the allocated memory

	switch priority {
	case 1:
		hostPriorities = append(highlightedRouters, hostPriorities...)
	case 2:
		speakerPriorities = append(highlightedRouters, speakerPriorities...)
	case 3:
		allPriorities = append(highlightedRouters, allPriorities...)
	case 4:
		nonePriorities = append(highlightedRouters, nonePriorities...)
	}

	sm.routers = append(sm.routers, hostPriorities...)
	sm.routers = append(sm.routers, speakerPriorities...)
	sm.routers = append(sm.routers, allPriorities...)
	sm.routers = append(sm.routers, nonePriorities...)

	sm.hasChangedRouters = true
	sm.mu.Unlock()

	// After reordering the routers, we need to trigger renegotiation for all sessions to ensure the new order is reflected on the client side
	for _, session := range sm.sessions {
		sm.SuscribeToPageinatedRouters(session, session.GetCurrentViewPage(), MAX_STREAMS_PER_PAGE) // Resubscribe to all routers to trigger renegotiation
	}
}
