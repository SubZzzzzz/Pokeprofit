package discord

import (
	"sync"
	"time"
)

// RateLimiter provides per-user rate limiting for Discord commands.
type RateLimiter struct {
	mu       sync.RWMutex
	limits   map[string]*commandLimits
	config   RateLimitConfig
	cleanupC chan struct{}
}

// commandLimits tracks rate limits for all commands for a single user.
type commandLimits struct {
	commands map[string]*userLimit
}

// userLimit tracks the last usage time and count for a user/command pair.
type userLimit struct {
	lastUsed time.Time
	count    int
	window   time.Time
}

// RateLimitConfig defines rate limits for each command.
type RateLimitConfig struct {
	// Analyze: 1 per user per 5 minutes (long-running operation)
	AnalyzeLimit    int
	AnalyzeWindow   time.Duration
	// Results: 5 per user per minute
	ResultsLimit    int
	ResultsWindow   time.Duration
	// Filter: 10 per user per minute
	FilterLimit     int
	FilterWindow    time.Duration
	// Default limits for unspecified commands
	DefaultLimit    int
	DefaultWindow   time.Duration
	// Cleanup interval for expired entries
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig returns the default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AnalyzeLimit:    1,
		AnalyzeWindow:   5 * time.Minute,
		ResultsLimit:    5,
		ResultsWindow:   1 * time.Minute,
		FilterLimit:     10,
		FilterWindow:    1 * time.Minute,
		DefaultLimit:    10,
		DefaultWindow:   1 * time.Minute,
		CleanupInterval: 10 * time.Minute,
	}
}

// NewRateLimiter creates a new RateLimiter with the given configuration.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		limits:   make(map[string]*commandLimits),
		config:   cfg,
		cleanupC: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// NewDefaultRateLimiter creates a new RateLimiter with default configuration.
func NewDefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(DefaultRateLimitConfig())
}

// Allow checks if a command is allowed for a user and records the usage if allowed.
// Returns true if the command is allowed, false if rate limited.
func (rl *RateLimiter) Allow(userID, command string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, window := rl.getCommandLimits(command)
	now := time.Now()

	// Get or create user limits
	userLimits, ok := rl.limits[userID]
	if !ok {
		userLimits = &commandLimits{
			commands: make(map[string]*userLimit),
		}
		rl.limits[userID] = userLimits
	}

	// Get or create command limit
	cmdLimit, ok := userLimits.commands[command]
	if !ok {
		cmdLimit = &userLimit{
			lastUsed: now,
			count:    1,
			window:   now,
		}
		userLimits.commands[command] = cmdLimit
		return true
	}

	// Check if we're in a new window
	if now.Sub(cmdLimit.window) >= window {
		// Reset the window
		cmdLimit.window = now
		cmdLimit.count = 1
		cmdLimit.lastUsed = now
		return true
	}

	// Check if under limit
	if cmdLimit.count < limit {
		cmdLimit.count++
		cmdLimit.lastUsed = now
		return true
	}

	// Rate limited
	return false
}

// Check checks if a command would be allowed without recording the usage.
func (rl *RateLimiter) Check(userID, command string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, window := rl.getCommandLimits(command)
	now := time.Now()

	userLimits, ok := rl.limits[userID]
	if !ok {
		return true
	}

	cmdLimit, ok := userLimits.commands[command]
	if !ok {
		return true
	}

	// Check if we're in a new window
	if now.Sub(cmdLimit.window) >= window {
		return true
	}

	return cmdLimit.count < limit
}

// TimeUntilAllowed returns the duration until the next allowed usage.
// Returns 0 if usage is currently allowed.
func (rl *RateLimiter) TimeUntilAllowed(userID, command string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	_, window := rl.getCommandLimits(command)
	now := time.Now()

	userLimits, ok := rl.limits[userID]
	if !ok {
		return 0
	}

	cmdLimit, ok := userLimits.commands[command]
	if !ok {
		return 0
	}

	// Check if we're in a new window
	windowEnd := cmdLimit.window.Add(window)
	if now.After(windowEnd) {
		return 0
	}

	// Check if under limit
	limit, _ := rl.getCommandLimits(command)
	if cmdLimit.count < limit {
		return 0
	}

	return windowEnd.Sub(now)
}

// Reset resets the rate limit for a specific user/command pair.
func (rl *RateLimiter) Reset(userID, command string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	userLimits, ok := rl.limits[userID]
	if !ok {
		return
	}

	delete(userLimits.commands, command)
}

// ResetUser resets all rate limits for a user.
func (rl *RateLimiter) ResetUser(userID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.limits, userID)
}

// getCommandLimits returns the limit and window for a command.
func (rl *RateLimiter) getCommandLimits(command string) (int, time.Duration) {
	switch command {
	case "analyze":
		return rl.config.AnalyzeLimit, rl.config.AnalyzeWindow
	case "results":
		return rl.config.ResultsLimit, rl.config.ResultsWindow
	case "filter":
		return rl.config.FilterLimit, rl.config.FilterWindow
	default:
		return rl.config.DefaultLimit, rl.config.DefaultWindow
	}
}

// cleanupLoop periodically removes expired rate limit entries.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.cleanupC:
			return
		}
	}
}

// cleanup removes expired entries.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	maxWindow := rl.config.AnalyzeWindow // Longest window
	if rl.config.ResultsWindow > maxWindow {
		maxWindow = rl.config.ResultsWindow
	}
	if rl.config.FilterWindow > maxWindow {
		maxWindow = rl.config.FilterWindow
	}
	if rl.config.DefaultWindow > maxWindow {
		maxWindow = rl.config.DefaultWindow
	}

	for userID, userLimits := range rl.limits {
		for cmd, cmdLimit := range userLimits.commands {
			if now.Sub(cmdLimit.lastUsed) > maxWindow*2 {
				delete(userLimits.commands, cmd)
			}
		}

		// Remove user if no commands left
		if len(userLimits.commands) == 0 {
			delete(rl.limits, userID)
		}
	}
}

// Stop stops the rate limiter and cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.cleanupC)
}

// Stats returns current rate limiter statistics.
type RateLimiterStats struct {
	TotalUsers    int
	TotalCommands int
}

// Stats returns the current rate limiter statistics.
func (rl *RateLimiter) Stats() RateLimiterStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	totalCommands := 0
	for _, userLimits := range rl.limits {
		totalCommands += len(userLimits.commands)
	}

	return RateLimiterStats{
		TotalUsers:    len(rl.limits),
		TotalCommands: totalCommands,
	}
}
