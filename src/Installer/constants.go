/*
 * SPDX-License-Identifier: GPL-3.0
 * Syncord Installer, a cross platform gui/cli app for installing Syncord
 * Copyright (c) 2023 Vendicated and SYNCORD contributors
 */

package main

import (
	"image/color"
	"vencordinstaller/buildinfo"
)

const VersionURL = "https://raw.githubusercontent.com/Ryze113/Syncord/main/build/version.txt"
const SyncordDistURL = "https://raw.githubusercontent.com/Ryze113/Syncord/main/build/"
const InstallerReleaseUrl = "https://api.github.com/repos/Vendicated/Installer/releases/latest"
const InstallerReleaseUrlFallback = "https://SYNCORD.dev/releases/installer"

var SyncordFiles = []string{"patcher.js", "preload.js", "renderer.js", "renderer.css"}

var UserAgent = "SyncordInstaller/" + buildinfo.InstallerGitHash + " (https://github.com/Ryze113/Syncord)"

var (
	DiscordGreen  = color.RGBA{R: 0x2D, G: 0x7C, B: 0x46, A: 0xFF}
	DiscordRed    = color.RGBA{R: 0xEC, G: 0x41, B: 0x44, A: 0xFF}
	DiscordBlue   = color.RGBA{R: 0x58, G: 0x65, B: 0xF2, A: 0xFF}
	DiscordYellow = color.RGBA{R: 0xfe, G: 0xe7, B: 0x5c, A: 0xff}
	GrayBg        = color.RGBA{R: 0x2C, G: 0x2F, B: 0x33, A: 0xFF}
)

var LinuxDiscordNames = []string{
	"Discord",
	"DiscordPTB",
	"DiscordCanary",
	"DiscordDevelopment",
	"discord",
	"discordptb",
	"discordcanary",
	"discorddevelopment",
	"discord-ptb",
	"discord-canary",
	"discord-development",
	// Flatpak
	"com.discordapp.Discord",
	"com.discordapp.DiscordPTB",
	"com.discordapp.DiscordCanary",
	"com.discordapp.DiscordDevelopment",
}
