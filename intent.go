// Copyright (c) 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"fmt"

	"maunium.net/go/mautrix"
)

type IntentAPI struct {
	*mautrix.Client
	bot       *mautrix.Client
	as        *AppService
	Localpart string
	UserID    string
}

func (as *AppService) NewIntentAPI(localpart string) *IntentAPI {
	userID := fmt.Sprintf("@%s:%s", localpart, as.HomeserverDomain)
	bot := as.BotClient()
	if userID == bot.UserID {
		bot = nil
	}
	return &IntentAPI{
		Client:    as.Client(userID),
		bot:       bot,
		as:        as,
		Localpart: localpart,
		UserID:    userID,
	}
}

func (intent *IntentAPI) Register() error {
	_, _, err := intent.Client.Register(&mautrix.ReqRegister{
		Username: intent.Localpart,
	})
	if err != nil {
		return err
	}
	return nil
}

func (intent *IntentAPI) EnsureRegistered() error {
	if intent.as.StateStore.IsRegistered(intent.UserID) {
		return nil
	}

	err := intent.Register()
	httpErr, ok := err.(mautrix.HTTPError)
	if !ok || httpErr.RespError.ErrCode != "M_USER_IN_USE" {
		return err
	}
	intent.as.StateStore.MarkRegistered(intent.UserID)
	return nil
}

func (intent *IntentAPI) EnsureJoined(roomID string) error {
	if intent.as.StateStore.IsInRoom(roomID, intent.UserID) {
		return nil
	}

	if err := intent.EnsureRegistered(); err != nil {
		return err
	}

	resp, err := intent.JoinRoom(roomID, "", nil)
	if err != nil {
		httpErr, ok := err.(mautrix.HTTPError)
		if !ok || httpErr.RespError.ErrCode != "M_FORBIDDEN" || intent.bot == nil {
			return httpErr
		}
		_, inviteErr := intent.bot.InviteUser(roomID, &mautrix.ReqInviteUser{
			UserID: intent.UserID,
		})
		if inviteErr != nil {
			return err
		}
		resp, err = intent.JoinRoom(roomID, "", nil)
		if err != nil {
			return err
		}
	}
	intent.as.StateStore.SetMembership(resp.RoomID, intent.UserID, "join")
	return nil
}

func (intent *IntentAPI) SendMessageEvent(roomID string, eventType mautrix.EventType, contentJSON interface{}) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendMessageEvent(roomID, eventType, contentJSON)
}

func (intent *IntentAPI) SendMassagedMessageEvent(roomID string, eventType mautrix.EventType, contentJSON interface{}, ts int64) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendMessageEvent(roomID, eventType, contentJSON, mautrix.ReqSendEvent{Timestamp: ts})
}

func (intent *IntentAPI) SendStateEvent(roomID string, eventType mautrix.EventType, stateKey string, contentJSON interface{}) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendStateEvent(roomID, eventType, stateKey, contentJSON)
}

func (intent *IntentAPI) SendMassagedStateEvent(roomID string, eventType mautrix.EventType, stateKey string, contentJSON interface{}, ts int64) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendMassagedStateEvent(roomID, eventType, stateKey, contentJSON, ts)
}

func (intent *IntentAPI) StateEvent(roomID string, eventType mautrix.EventType, stateKey string, outContent interface{}) (err error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return err
	}
	return intent.Client.StateEvent(roomID, eventType, stateKey, outContent)
}

func (intent *IntentAPI) PowerLevels(roomID string) (pl *mautrix.PowerLevels, err error) {
	pl = intent.as.StateStore.GetPowerLevels(roomID)
	if pl == nil {
		pl = &mautrix.PowerLevels{}
		err = intent.StateEvent(roomID, mautrix.StatePowerLevels, "", pl)
		if err == nil {
			intent.as.StateStore.SetPowerLevels(roomID, pl)
		}
	}
	return
}

func (intent *IntentAPI) SetPowerLevels(roomID string, levels *mautrix.PowerLevels) (resp *mautrix.RespSendEvent, err error) {
	resp, err = intent.SendStateEvent(roomID, mautrix.StatePowerLevels, "", &levels)
	if err == nil {
		intent.as.StateStore.SetPowerLevels(roomID, levels)
	}
	return
}

func (intent *IntentAPI) SetPowerLevel(roomID, userID string, level int) (*mautrix.RespSendEvent, error) {
	pl, err := intent.PowerLevels(roomID)
	if err != nil {
		return nil, err
	}

	if pl.GetUserLevel(userID) != level {
		pl.SetUserLevel(userID, level)
		return intent.SendStateEvent(roomID, mautrix.StatePowerLevels, "", &pl)
	}
	return nil, nil
}

func (intent *IntentAPI) UserTyping(roomID string, typing bool, timeout int64) (resp *mautrix.RespTyping, err error) {
	if intent.as.StateStore.IsTyping(roomID, intent.UserID) == typing {
		return
	}
	resp, err = intent.Client.UserTyping(roomID, typing, timeout)
	if err != nil {
		return
	}
	if !typing {
		timeout = -1
	}
	intent.as.StateStore.SetTyping(roomID, intent.UserID, timeout)
	return
}

func (intent *IntentAPI) SendText(roomID, text string) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendText(roomID, text)
}

func (intent *IntentAPI) SendImage(roomID, body, url string) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendImage(roomID, body, url)
}

func (intent *IntentAPI) SendVideo(roomID, body, url string) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendVideo(roomID, body, url)
}

func (intent *IntentAPI) SendNotice(roomID, text string) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendNotice(roomID, text)
}

func (intent *IntentAPI) RedactEvent(roomID, eventID string, req ...mautrix.ReqRedact) (*mautrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.RedactEvent(roomID, eventID, req...)
}

func (intent *IntentAPI) SetRoomName(roomID, roomName string) (*mautrix.RespSendEvent, error) {
	return intent.SendStateEvent(roomID, mautrix.StateRoomName, "", map[string]interface{}{
		"name": roomName,
	})
}

func (intent *IntentAPI) SetRoomAvatar(roomID, avatarURL string) (*mautrix.RespSendEvent, error) {
	return intent.SendStateEvent(roomID, mautrix.StateRoomAvatar, "", map[string]interface{}{
		"url": avatarURL,
	})
}

func (intent *IntentAPI) SetRoomTopic(roomID, topic string) (*mautrix.RespSendEvent, error) {
	return intent.SendStateEvent(roomID, mautrix.StateTopic, "", map[string]interface{}{
		"topic": topic,
	})
}

func (intent *IntentAPI) SetDisplayName(displayName string) error {
	if err := intent.EnsureRegistered(); err != nil {
		return err
	}
	return intent.Client.SetDisplayName(displayName)
}

func (intent *IntentAPI) SetAvatarURL(avatarURL string) error {
	if err := intent.EnsureRegistered(); err != nil {
		return err
	}
	return intent.Client.SetAvatarURL(avatarURL)
}

func (intent *IntentAPI) EnsureInvited(roomID, userID string) error {
	if !intent.as.StateStore.IsInvited(roomID, userID) {
		_, err := intent.Client.InviteUser(roomID, &mautrix.ReqInviteUser{
			UserID: userID,
		})
		return err
	}
	return nil
}
