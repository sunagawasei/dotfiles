# dotfiles

Personal macOS development environment configuration files with a consistent Vercel Geist design system theme.

## Overview

This repository contains my personal dotfiles for macOS development tools. The configurations are manually managed without using stow or other dotfile managers. All configurations follow a unified dark theme based on the Vercel Geist design system.

## Main Components

### Terminal Emulators

#### Wezterm (Primary)
- **Config**: `wezterm/wezterm.lua`
- **Keybindings**: `wezterm/keybinds.lua`
- **Features**: GPU accelerated, multiplexing, custom Vercel dark theme
- **Leader key**: `Ctrl+q`

#### Ghostty (Secondary)
- **Config**: `ghostty/config`
- **Commands reference**: `ghostty/command.md`
- **Features**: Native macOS terminal, Vercel dark theme

### Text Editors

#### Neovim (LazyVim)
- **Config**: `nvim/init.lua` and `nvim/lua/`
- **Framework**: LazyVim (pre-configured Neovim setup)
- **Formatter**: Stylua configured in `nvim/stylua.toml`
- **Features**: Japanese IME support, performance optimizations

#### Zed
- **Config**: `zed/settings.json`
- **Keybindings**: `zed/keymap.json`
- **Themes**: `zed/themes/`

### Shell Configuration

#### Fish Shell
- **Config**: `fish/conf.d/`

#### Zsh
- **Config**: `shell/zsh/.zshrc`
- **Abbreviations**: `zsh-abbr/user-abbreviations`
- **Plugin manager**: Antigen

#### Starship Prompt
- **Config**: `starship.toml`
- **Theme**: Minimal with Vercel Geist colors
- **Features**: Git status, time display, custom prompt character (▲)

### Other Tools

#### Git
- **Config**: `git/gitconfig`
- **Global ignore**: `git/gitignore`

#### tmux
- **Config**: `tmux/tmux.conf`

#### LazyGit
- **Config**: `lazygit/config.yml`

#### Raycast Extensions
- **Location**: `raycast/extensions/`
- **Custom extensions**: Multiple productivity tools and integrations

## Color System

The entire configuration uses a consistent color scheme based on Vercel's Geist design system. See `vercel-geist-colors.md` for the complete color palette reference including:
- 10-step color scales for Gray, Blue, Red, Amber, Green, Teal, Purple, and Pink
- Consistent background, foreground, and accent colors across all applications

## Setup

As this is a manually managed dotfiles repository:

1. Clone this repository to `~/.config/`
2. Each application's configuration files should be symlinked or copied to their expected locations
3. No automated setup scripts are provided - configure each tool individually

### Common Configuration Locations

- Neovim: `~/.config/nvim/`
- Wezterm: `~/.config/wezterm/`
- Ghostty: `~/.config/ghostty/`
- Fish: `~/.config/fish/`
- Starship: `~/.config/starship.toml`
- Git: `~/.gitconfig` (symlink from `~/.config/git/gitconfig`)

## Features

- **Unified Theme**: Consistent Vercel Geist dark theme across all tools
- **Japanese Language Support**: Full Japanese input method support in Neovim and terminals
- **Performance Optimized**: Configurations tuned for fast startup and responsive interaction
- **Modular Structure**: Each tool's configuration is self-contained
- **Version Controlled**: Git managed with descriptive commit messages

## Directory Structure

```
.config/
├── nvim/               # Neovim configuration (LazyVim)
├── wezterm/            # Wezterm terminal configuration
├── ghostty/            # Ghostty terminal configuration
├── fish/               # Fish shell configuration
├── shell/              # Shell utilities (zsh configs)
├── git/                # Git configuration
├── tmux/               # tmux configuration
├── lazygit/            # LazyGit configuration
├── starship.toml       # Starship prompt configuration
├── raycast/            # Raycast extensions and config
├── zed/                # Zed editor configuration
├── karabiner/          # Karabiner-Elements (keyboard customization)
├── dlv/                # Go Delve debugger config
├── CLAUDE.md           # Claude AI assistant instructions
└── vercel-geist-colors.md  # Color system documentation
```

## Notes

- No remote repository configured (local only)
- Backup files (`*.bak`) are automatically created by some tools
- Configuration files contain Japanese comments and documentation
- All commit messages follow either Japanese or English conventions