package hypeserver

import (
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
)

func GetTotalUsers() uint64 {
	users, success := user.GetTotalServed(DB)
	if success != true {
		return TotalUsers
	}
	return users
}

func GetTotalGuilds() uint64 {
	guilds, success := guild.GetTotalServed(DB)
	if success != true {
		return TotalServers
	}
	return guilds
}

func GetStats() (uint64, uint64) {
	TotalServers = GetTotalGuilds()
	TotalUsers = GetTotalUsers()

	return TotalServers, TotalUsers
}
