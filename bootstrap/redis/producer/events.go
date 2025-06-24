// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package producer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/bootstrap"
)

const (
	configPrefix        = "config."
	configCreate        = configPrefix + "create"
	configUpdate        = configPrefix + "update"
	configRemove        = configPrefix + "remove"
	configList          = configPrefix + "list"
	configHandlerRemove = configPrefix + "remove_handler"

	thingPrefix            = "thing."
	thingBootstrap         = thingPrefix + "bootstrap"
	thingStateChange       = thingPrefix + "change_state"
	thingUpdateConnections = thingPrefix + "update_connections"
	thingDisconnect        = thingPrefix + "disconnect"

	channelPrefix        = "channel."
	channelHandlerRemove = channelPrefix + "remove_handler"
	channelUpdateHandler = channelPrefix + "update_handler"

	certUpdate = "cert.update"
)

type event interface {
	encode() (map[string]interface{}, error)
}

var (
	_ event = (*configEvent)(nil)
	_ event = (*removeConfigEvent)(nil)
	_ event = (*bootstrapEvent)(nil)
	_ event = (*changeStateEvent)(nil)
	_ event = (*updateConnectionsEvent)(nil)
	_ event = (*updateCertEvent)(nil)
	_ event = (*listConfigsEvent)(nil)
	_ event = (*removeHandlerEvent)(nil)
)

type configEvent struct {
	bootstrap.Config
	operation string
}

func (ce configEvent) encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"state":     ce.State.String(),
		"operation": ce.operation,
	}
	if ce.ThingID != "" {
		val["thing_id"] = ce.ThingID
	}
	if ce.Content != "" {
		val["content"] = ce.Content
	}
	if ce.Owner != "" {
		val["owner"] = ce.Owner
	}
	if ce.Name != "" {
		val["name"] = ce.Name
	}
	if ce.ExternalID != "" {
		val["external_id"] = ce.ExternalID
	}
	if len(ce.Channels) > 0 {
		channels := make([]string, len(ce.Channels))
		for i, ch := range ce.Channels {
			channels[i] = ch.ID
		}
		val["channels"] = fmt.Sprintf("[%s]", strings.Join(channels, ", "))
	}
	if ce.ClientCert != "" {
		val["client_cert"] = ce.ClientCert
	}
	if ce.ClientKey != "" {
		val["client_key"] = ce.ClientKey
	}
	if ce.CACert != "" {
		val["ca_cert"] = ce.CACert
	}
	if ce.Content != "" {
		val["content"] = ce.Content
	}

	return val, nil
}

type removeConfigEvent struct {
	mfThing string
}

func (rce removeConfigEvent) encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"thing_id":  rce.mfThing,
		"operation": configRemove,
	}, nil
}

type listConfigsEvent struct {
	offset       uint64
	limit        uint64
	fullMatch    map[string]string
	partialMatch map[string]string
}

func (rce listConfigsEvent) encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"offset":    rce.offset,
		"limit":     rce.limit,
		"operation": configList,
	}
	if len(rce.fullMatch) > 0 {
		data, err := json.Marshal(rce.fullMatch)
		if err != nil {
			return map[string]interface{}{}, err
		}

		val["full_match"] = data
	}

	if len(rce.partialMatch) > 0 {
		data, err := json.Marshal(rce.partialMatch)
		if err != nil {
			return map[string]interface{}{}, err
		}

		val["full_match"] = data
	}
	return val, nil
}

type bootstrapEvent struct {
	bootstrap.Config
	externalID string
	success    bool
}

func (be bootstrapEvent) encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"external_id": be.externalID,
		"success":     be.success,
		"operation":   thingBootstrap,
	}

	if be.ThingID != "" {
		val["thing_id"] = be.ThingID
	}
	if be.Content != "" {
		val["content"] = be.Content
	}
	if be.Owner != "" {
		val["owner"] = be.Owner
	}
	if be.Name != "" {
		val["name"] = be.Name
	}
	if be.ExternalID != "" {
		val["external_id"] = be.ExternalID
	}
	if len(be.Channels) > 0 {
		channels := make([]string, len(be.Channels))
		for i, ch := range be.Channels {
			channels[i] = ch.ID
		}
		val["channels"] = fmt.Sprintf("[%s]", strings.Join(channels, ", "))
	}
	if be.ClientCert != "" {
		val["client_cert"] = be.ClientCert
	}
	if be.ClientKey != "" {
		val["client_key"] = be.ClientKey
	}
	if be.CACert != "" {
		val["ca_cert"] = be.CACert
	}
	if be.Content != "" {
		val["content"] = be.Content
	}
	return val, nil
}

type changeStateEvent struct {
	mfThing string
	state   bootstrap.State
}

func (cse changeStateEvent) encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"thing_id":  cse.mfThing,
		"state":     cse.state.String(),
		"operation": thingStateChange,
	}, nil
}

type updateConnectionsEvent struct {
	mfThing    string
	mfChannels []string
}

func (uce updateConnectionsEvent) encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"thing_id":  uce.mfThing,
		"channels":  fmt.Sprintf("[%s]", strings.Join(uce.mfChannels, ", ")),
		"operation": thingUpdateConnections,
	}, nil
}

type updateCertEvent struct {
	thingKey, clientCert, clientKey, caCert string
}

func (uce updateCertEvent) encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"thing_key":   uce.thingKey,
		"client_cert": uce.clientCert,
		"client_key":  uce.clientKey,
		"ca_cert":     uce.caCert,
		"operation":   certUpdate,
	}, nil
}

type removeHandlerEvent struct {
	id        string
	operation string
}

func (rhe removeHandlerEvent) encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"config_id": rhe.id,
		"operation": rhe.operation,
	}, nil
}

type updateChannelHandlerEvent struct {
	bootstrap.Channel
}

func (uche updateChannelHandlerEvent) encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": channelUpdateHandler,
	}

	if uche.ID != "" {
		val["channel_id"] = uche.ID
	}
	if uche.Name != "" {
		val["name"] = uche.Name
	}
	if uche.Metadata != nil {
		metadata, err := json.Marshal(uche.Metadata)
		if err != nil {
			return map[string]interface{}{}, err
		}

		val["metadata"] = metadata
	}
	return val, nil
}

type disconnectThingEvent struct {
	thingID   string
	channelID string
}

func (dte disconnectThingEvent) encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"thing_id":   dte.thingID,
		"channel_id": dte.channelID,
		"operation":  thingDisconnect,
	}, nil
}
