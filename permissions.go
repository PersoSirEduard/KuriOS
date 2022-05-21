package main

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HasPermission(session *discordgo.Session, member *discordgo.Member, guildID string, permission string) bool {
	guild := GetGuild(session, guildID)

	// Check general permissions for everyone
	if _, ok := systemPermissions["everyone"]; ok {
		for _, perm := range systemPermissions["everyone"] {
			if perm == permission {
				return true
			} else if perm == "*" {
				return true
			}
		}
	}

	// Check for role permissions
	for _, roleId := range (*member).Roles {
		roleName := GetRoleName(guild, roleId)

		if _, ok := systemPermissions[roleName]; ok {

			for _, perm := range systemPermissions[roleName] {
				if perm == permission {
					return true
				} else if perm == "*" {
					return true
				}
			}
		}
	}

	return false
}

func GetRoleName(guild *discordgo.Guild, roleID string) string {
	for _, role := range (*guild).Roles {
		if role.ID == roleID {
			return role.Name
		}
	}
	return ""
}

func GetRoleId(guild *discordgo.Guild, roleName string) (string, error) {
	for _, role := range (*guild).Roles {
		if role.Name == roleName {
			return role.ID, nil
		}
	}
	return "", errors.New("Error: Role \" " + roleName + " \" not found.")
}

func GetGuild(session *discordgo.Session, guildID string) *discordgo.Guild {
	for _, guild := range (*session).State.Guilds {
		if guild.ID == guildID {
			return guild
		}
	}
	return nil
}

func GetRolePriority(roleName string) int {
	for i, role := range priorityPermissions {
		if role == roleName {
			return i
		}
	}
	return -1
}

func GetUserHighestPriorityRole(session *discordgo.Session, member *discordgo.Member, guildID string) string {
	guild := GetGuild(session, guildID)
	highestPriority := -1
	highestPriorityRole := ""

	for _, roleId := range (*member).Roles {
		roleName := GetRoleName(guild, roleId)

		if GetRolePriority(roleName) > highestPriority {
			highestPriority = GetRolePriority(roleName)
			highestPriorityRole = roleName
		}
	}

	return highestPriorityRole
}

func HasRole(session *discordgo.Session, member *discordgo.Member, guildID string, roleName string) bool {
	for _, role := range (*member).Roles {
		if GetRoleName(GetGuild(session, guildID), role) == roleName {
			return true
		}
	}
	return false
}

func subscribeToRole(session *discordgo.Session, member *discordgo.Member, guildID string, userID string, roleName string) error {

	newRolePriority := GetRolePriority(roleName)
	if newRolePriority == -1 {
		return errors.New("Error: Role \"" + roleName + "\" not found.")
	}

	memberCurrentRole := GetUserHighestPriorityRole(session, member, guildID)
	if memberCurrentRole == "" {
		// Get everyone role
		memberCurrentRole = "everyone"
	}

	if newRolePriority >= GetRolePriority(memberCurrentRole) {
		return errors.New("Error: You do not have permission to subscribe to this role.")
	}

	// Subscribe to role if not already subscribed
	if HasRole(session, member, guildID, roleName) {
		return errors.New("Error: You are already subscribed to this role.")
	}

	// Subscribe to role
	roleId, err := GetRoleId(GetGuild(session, guildID), roleName)
	if err != nil {
		return err
	}

	err = (*session).GuildMemberRoleAdd(guildID, userID, roleId)

	return err
}

func unsubscribeFromRole(session *discordgo.Session, member *discordgo.Member, guildID string, userID string, roleName string) error {

	// Check if role exists
	if !HasRole(session, member, guildID, roleName) {
		return errors.New("Error: You are not subscribed to this role.")
	}

	// Unsubscribe from role
	roleId, err := GetRoleId(GetGuild(session, guildID), roleName)
	if err != nil {
		return err
	}

	err = (*session).GuildMemberRoleRemove(guildID, userID, roleId)
	return err

}

func IsTimeAvailable(startDate string, endDate string) bool {

	// Handle "*" for startDate. Current time minus 10 minutes
	if startDate == "*" {
		startDate = time.Now().Add(-10 * time.Minute).Format("2006-01-02 15:04:05")
	}

	// Handle "*" for endDate. Current time minus 10 minutes
	if endDate == "*" {
		endDate = time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05")
	}

	// Get unix time for start and end date
	startTime, err := time.Parse("2006-01-02 15:04:05", startDate)
	if err != nil {
		return false
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", endDate)
	if err != nil {
		return false
	}

	now, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

	// Compare times
	return now.Unix() >= startTime.Unix() && now.Unix() <= endTime.Unix()
}
