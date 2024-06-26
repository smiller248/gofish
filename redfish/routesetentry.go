//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"encoding/json"
	"reflect"

	"github.com/stmcginnis/gofish/common"
)

// RouteSetEntry This Resource contains the content of a route set in the Redfish Specification.
type RouteSetEntry struct {
	common.Entity
	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataEtag is the odata etag.
	ODataEtag string `json:"@odata.etag"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// Description provides a description of this resource.
	Description string
	// EgressIdentifier shall contain the interface identifier corresponding to this route.
	EgressIdentifier int
	// HopCount shall contain the number of hops to the destination component from the indicated egress interface.
	HopCount int
	// Oem shall contain the OEM extensions. All values for properties that this object contains shall conform to the
	// Redfish Specification-described requirements.
	OEM json.RawMessage `json:"Oem"`
	// VCAction shall contain the index to the VCAT entry corresponding to this route.
	VCAction int
	// Valid shall indicate whether the entry is valid.
	Valid bool
	// rawData holds the original serialized JSON so we can compare updates.
	rawData []byte
}

// UnmarshalJSON unmarshals a RouteSetEntry object from the raw JSON.
func (routesetentry *RouteSetEntry) UnmarshalJSON(b []byte) error {
	type temp RouteSetEntry
	var t struct {
		temp
	}

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	*routesetentry = RouteSetEntry(t.temp)

	// Extract the links to other entities for later

	// This is a read/write object, so we need to save the raw object data for later
	routesetentry.rawData = b

	return nil
}

// Update commits updates to this object's properties to the running system.
func (routesetentry *RouteSetEntry) Update() error {
	// Get a representation of the object's original state so we can find what
	// to update.
	original := new(RouteSetEntry)
	original.UnmarshalJSON(routesetentry.rawData)

	readWriteFields := []string{
		"EgressIdentifier",
		"HopCount",
		"VCAction",
		"Valid",
	}

	originalElement := reflect.ValueOf(original).Elem()
	currentElement := reflect.ValueOf(routesetentry).Elem()

	return routesetentry.Entity.Update(originalElement, currentElement, readWriteFields)
}

// GetRouteSetEntry will get a RouteSetEntry instance from the service.
func GetRouteSetEntry(c common.Client, uri string) (*RouteSetEntry, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routesetentry RouteSetEntry
	err = json.NewDecoder(resp.Body).Decode(&routesetentry)
	if err != nil {
		return nil, err
	}

	routesetentry.SetClient(c)
	return &routesetentry, nil
}

// ListReferencedRouteSetEntrys gets the collection of RouteSetEntry from
// a provided reference.
func ListReferencedRouteSetEntrys(c common.Client, link string) ([]*RouteSetEntry, error) {
	var result []*RouteSetEntry
	if link == "" {
		return result, nil
	}

	type GetResult struct {
		Item  *RouteSetEntry
		Link  string
		Error error
	}

	ch := make(chan GetResult)
	collectionError := common.NewCollectionError()
	get := func(link string) {
		routesetentry, err := GetRouteSetEntry(c, link)
		ch <- GetResult{Item: routesetentry, Link: link, Error: err}
	}

	go func() {
		err := common.CollectList(get, c, link)
		if err != nil {
			collectionError.Failures[link] = err
		}
		close(ch)
	}()

	for r := range ch {
		if r.Error != nil {
			collectionError.Failures[r.Link] = r.Error
		} else {
			result = append(result, r.Item)
		}
	}

	if collectionError.Empty() {
		return result, nil
	}

	return result, collectionError
}
