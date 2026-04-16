package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (server *Server) handleGeoadd(arr []string) string {
	if len(arr) < 5 {
		return "-ERR Invalid input. Please try: GEOADD <set_key> <latitude> <longitude> <location>\r\n"
	}

	setKey := arr[1]
	longitude, err := strconv.ParseFloat(arr[2], 64)
	if err != nil {
		return "-ERR Invalid longitude\r\n"
	}

	latitude, err := strconv.ParseFloat(arr[3], 64)
	if err != nil {
		return "-ERR Invalid latitude\r\n"
	}

	if longitude < MIN_LONGITUDE || longitude > MAX_LONGITUDE || latitude < MIN_LATITUDE || latitude > MAX_LATITUDE {
		return fmt.Sprintf("-ERR invalid longitude,latitude pair, %f,%f\r\n", longitude, latitude)
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	set, prs := server.zsetsMap[setKey]
	if !prs {
		set = NewSortedSet()
		server.zsetsMap[setKey] = set
	}

	score := encode(&Coordinates{
		Longitude: longitude,
		Latitude:  latitude,
	})

	set.AddOrUpdate(arr[4], float64(score))
	return ":1\r\n"
}

func (server *Server) handleGeopos(arr []string) string {
	if len(arr) < 3 {
		return "-ERR Invalid input. Please try: GEOPOS <zset_key> <location>"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	locations := arr[2:]

	var sb strings.Builder
	addRespArrayHeader(&sb, len(locations))

	set, exists := server.zsetsMap[arr[1]]
	if !exists {
		for range locations {
			sb.WriteString("*-1\r\n")
		}
		return sb.String()
	}

	for _, location := range locations {
		score, exists := set.GetMemberScore(location)
		if !exists {
			sb.WriteString("*-1\r\n")
			continue
		}

		coordinates := decode(uint64(score))
		addRespArrayHeader(&sb, 2)
		addRespString(&sb, strconv.FormatFloat(coordinates.Longitude, 'f', -1, 64))
		addRespString(&sb, strconv.FormatFloat(coordinates.Latitude, 'f', -1, 64))
	}

	return sb.String()
}

func (server *Server) handleGeodist(arr []string) string {
	if len(arr) < 4 {
		return "-ERR Invalid syntax. Please try: GEOADD <zset_key> <location_1> <location_2>\r\n"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	zsetKey, locationName1, locationName2 := arr[1], arr[2], arr[3]
	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return "$-1\r\n"
	}

	location1, exists1 := set.ElemDict[locationName1]
	location2, exists2 := set.ElemDict[locationName2]
	if !exists1 || !exists2 {
		return "$-1\r\n"
	}

	coord1 := decode(uint64(location1.Score))
	coord2 := decode(uint64(location2.Score))

	rad1 := degPos(coord1.Latitude, coord1.Longitude)
	rad2 := degPos(coord2.Latitude, coord2.Longitude)

	var sb strings.Builder
	dist := hsDist(rad1, rad2)
	addRespString(&sb, strconv.FormatFloat(dist, 'f', -1, 64))
	return sb.String()
}

func (server *Server) handleGeosearch(arr []string) string {
	if len(arr) < 8 {
		return "-ERR Invalid input. Please try: GEOSEARCH <zset_key> FROMLONLAT <longitude> <latitude> BYRADIUS 100 m"
	}

	zsetKey := arr[1]
	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return "$-1\r\n"
	}

	longitude, err := strconv.ParseFloat(arr[3], 64)
	if err != nil {
		return "-ERR Invalid longitude\r\n"
	}

	latitude, err := strconv.ParseFloat(arr[4], 64)
	if err != nil {
		return "-ERR Invalid latitude\r\n"
	}

	radius, err := strconv.ParseFloat(arr[6], 64)
	unit := arr[7]
	radius = convertRadius(radius, unit)

	var sb strings.Builder
	var locations []string

	rad1 := degPos(latitude, longitude)
	for _, location := range set.ElemSlice {
		coord1 := decode(uint64(location.Score))
		rad2 := degPos(coord1.Latitude, coord1.Longitude)

		dist := hsDist(rad1, rad2)
		if dist < radius {
			locations = append(locations, location.Member)
		}
	}

	addRespArrayHeader(&sb, len(locations))
	for _, location := range locations {
		addRespString(&sb, location)
	}

	return sb.String()

}

func convertRadius(radius float64, unit string) float64 {
	switch unit {
	case "km":
		return float64(radius * 1000)
	case "mi":
		return 1609.344 * float64(radius)
	case "ft":
		return 0.3048 * float64(radius)
	}
	return float64(radius)
}
