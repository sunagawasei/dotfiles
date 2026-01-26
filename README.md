# dotfiles

Personal macOS development environment configuration files with a consistent custom dark theme.

## Overview

This repository contains my personal dotfiles for macOS development tools. The configurations are manually managed without using stow or other dotfile managers. All configurations follow a unified custom dark theme.

## Main Components

### Terminal Emulator

#### WezTerm
- **Config**: `wezterm/wezterm.lua`
- **Keybindings**: `wezterm/keybinds.lua`
- **Features**: GPU accelerated, multiplexing, custom dark theme
- **Leader key**: `Ctrl+q`

### Text Editors

#### Neovim (LazyVim)
- **Config**: `nvim/init.lua` and `nvim/lua/`
- **Framework**: LazyVim (pre-configured Neovim setup)
- **Formatter**: Stylua configured in `nvim/stylua.toml`
- **Features**: Japanese IME support, performance optimizations, AI integration (Copilot)
- **Documentation**: `nvim/CLAUDE.md` for detailed configuration

#### Vim (Legacy)
- **Config**: `vim/vimrc`
- **Usage**: Fallback for environments without Neovim

### Shell Configuration

#### Zsh
- **Config**: `zsh/.zshrc`
- **Features**: Comprehensive completion settings, aliases

#### Starship Prompt
- **Config**: `starship.toml`
- **Theme**: Minimal with custom colors
- **Features**: Git status, time display, custom prompt character (▲)

### Development Tools

#### Git
- **Config**: `git/gitconfig`
- **Global ignore**: `git/gitignore`, `git/ignore`

#### LazyGit
- **Config**: `lazygit/config.yml`
- **Features**: Git TUI with custom keybindings

#### Claude Code
- **Config**: `claude/settings.json` (via `$CLAUDE_CONFIG_DIR`)
- **Scripts**: `claude/statusline.sh` for terminal status line integration
- **Features**: AI-assisted development environment

#### Raycast
- **Config**: `raycast/config.json`
- **Features**: macOS productivity launcher

## Color System

The entire configuration uses a consistent custom color scheme. See `COLOR-SYSTEM.md` for the complete color palette reference including:
- Monochrome gradations and accent colors
- Consistent background, foreground, and accent colors across all applications

## Documentation

- **CLAUDE.md**: Instructions for Claude Code AI assistant
- **SHORTCUTS.md**: Comprehensive keymaps for Neovim, WezTerm, Zsh, and LazyGit

## Setup

As this is a manually managed dotfiles repository:

1. Clone this repository to `~/.config/`
2. Each application's configuration files should be symlinked or copied to their expected locations
3. No automated setup scripts are provided - configure each tool individually

### Common Configuration Locations

- Neovim: `~/.config/nvim/`
- WezTerm: `~/.config/wezterm/`
- Zsh: `~/.zshrc` (symlink from `zsh/.zshrc`)
- Starship: `~/.config/starship.toml`
- Git: `~/.gitconfig` (symlink from `git/gitconfig`)
- LazyGit: `~/.config/lazygit/`
- Claude Code: `$CLAUDE_CONFIG_DIR` (default: `~/.config/claude/`)

## Features

- **Unified Theme**: Consistent custom dark theme across all tools
- **Japanese Language Support**: Full Japanese input method support in Neovim and terminals
- **Performance Optimized**: Configurations tuned for fast startup and responsive interaction
- **Modular Structure**: Each tool's configuration is self-contained
- **AI Integration**: Claude Code and GitHub Copilot support
- **Version Controlled**: Git managed with descriptive commit messages

## Directory Structure

```
dotfiles/
├── nvim/               # Neovim configuration (LazyVim)
├── wezterm/            # WezTerm terminal configuration
├── vim/                # Vim configuration (legacy)
├── zsh/                # Zsh shell configuration
├── git/                # Git configuration
├── lazygit/            # LazyGit TUI configuration
├── claude/             # Claude Code AI settings
├── raycast/            # Raycast launcher configuration
├── starship.toml       # Starship prompt configuration
├── CLAUDE.md           # Claude AI assistant instructions
├── SHORTCUTS.md        # Keyboard shortcuts reference
└── COLOR-SYSTEM.md     # Color system documentation
```

## Notes

- Remote repository: `git@github.com:sunagawasei/dotfiles.git`
- Configuration files contain Japanese comments and documentation
- Commit messages follow either Japanese or English conventions
