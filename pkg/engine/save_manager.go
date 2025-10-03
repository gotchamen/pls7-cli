package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SaveManager handles all file system operations for game save/load functionality.
// It provides methods for saving games, loading games, listing saves, and managing
// the save directory structure.
type SaveManager struct {
	// SaveDir is the directory where save files are stored.
	SaveDir string
}

// SaveFileInfo contains metadata about a save file.
type SaveFileInfo struct {
	// Filename is the name of the save file.
	Filename string
	// FullPath is the complete path to the save file.
	FullPath string
	// CreatedAt is when the save file was created.
	CreatedAt time.Time
	// Size is the size of the save file in bytes.
	Size int64
	// GameMetadata contains basic game information from the save file.
	GameMetadata *GameMetadata
}

// NewSaveManager creates a new SaveManager with the specified save directory.
// If the directory doesn't exist, it will be created.
func NewSaveManager(saveDir string) (*SaveManager, error) {
	// Create save directory if it doesn't exist
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory %s: %w", saveDir, err)
	}

	return &SaveManager{
		SaveDir: saveDir,
	}, nil
}

// SaveGame saves the current game state to a file with the specified name.
// If filename is empty, it will generate a timestamp-based filename automatically.
// The filename should not include the .json extension as it will be added automatically.
func (sm *SaveManager) SaveGame(game *Game, filename string) error {
	// Generate timestamp-based filename if not provided
	if filename == "" {
		filename = fmt.Sprintf("save_%s", time.Now().Format("20060102_150405"))
	}

	// Sanitize filename
	filename = sm.sanitizeFilename(filename)
	if filename == "" {
		return fmt.Errorf("invalid filename after sanitization")
	}

	// Add .json extension if not present
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	// Create full path
	fullPath := filepath.Join(sm.SaveDir, filename)

	// Convert game to save data
	saveData := game.ToSaveData()

	// Serialize to JSON
	jsonData, err := saveData.SaveToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize game data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write save file %s: %w", fullPath, err)
	}

	logrus.Infof("Game saved successfully to %s", fullPath)
	return nil
}

// LoadGame loads a game from the specified save file.
// If filename is empty, it will load the most recently created save file.
func (sm *SaveManager) LoadGame(filename string) (*Game, error) {
	// If no filename provided, find the most recent save file
	if filename == "" {
		saves, err := sm.ListSaves()
		if err != nil {
			return nil, fmt.Errorf("failed to list save files: %w", err)
		}

		if len(saves) == 0 {
			return nil, fmt.Errorf("no save files found in directory: %s", sm.SaveDir)
		}

		// Use the most recent save file (ListSaves returns sorted by creation time, newest first)
		filename = saves[0].Filename
		logrus.Infof("Auto-loading most recent save file: %s", filename)
	}

	// Add .json extension if not present
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	// Create full path
	fullPath := filepath.Join(sm.SaveDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("save file %s does not exist", fullPath)
	}

	// Read file
	jsonData, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file %s: %w", fullPath, err)
	}

	// Deserialize from JSON
	saveData, err := LoadFromJSON(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse save file %s: %w", fullPath, err)
	}

	// Convert to game
	game, err := FromSaveData(saveData)
	if err != nil {
		return nil, fmt.Errorf("failed to restore game from save data: %w", err)
	}

	logrus.Infof("Game loaded successfully from %s", fullPath)
	return game, nil
}

// ListSaves returns a list of all save files in the save directory.
// The list is sorted by creation time (newest first).
func (sm *SaveManager) ListSaves() ([]SaveFileInfo, error) {
	var saves []SaveFileInfo

	// Read directory
	entries, err := os.ReadDir(sm.SaveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory %s: %w", sm.SaveDir, err)
	}

	// Process each .json file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		fullPath := filepath.Join(sm.SaveDir, entry.Name())

		// Get file info
		fileInfo, err := entry.Info()
		if err != nil {
			logrus.Warnf("Failed to get file info for %s: %v", entry.Name(), err)
			continue
		}

		// Try to read game metadata
		gameMetadata, err := sm.readGameMetadata(fullPath)
		if err != nil {
			logrus.Warnf("Failed to read metadata from %s: %v", entry.Name(), err)
			// Continue without metadata rather than failing completely
		}

		saveInfo := SaveFileInfo{
			Filename:     entry.Name(),
			FullPath:     fullPath,
			CreatedAt:    fileInfo.ModTime(),
			Size:         fileInfo.Size(),
			GameMetadata: gameMetadata,
		}

		saves = append(saves, saveInfo)
	}

	// Sort by creation time (newest first)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].CreatedAt.After(saves[j].CreatedAt)
	})

	return saves, nil
}

// DeleteSave removes a save file from the save directory.
func (sm *SaveManager) DeleteSave(filename string) error {
	// Add .json extension if not present
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	// Create full path
	fullPath := filepath.Join(sm.SaveDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("save file %s does not exist", fullPath)
	}

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete save file %s: %w", fullPath, err)
	}

	logrus.Infof("Save file %s deleted successfully", fullPath)
	return nil
}

// ValidateSaveFile checks if a save file is valid and can be loaded.
func (sm *SaveManager) ValidateSaveFile(filename string) error {
	// Add .json extension if not present
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	// Create full path
	fullPath := filepath.Join(sm.SaveDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("save file %s does not exist", fullPath)
	}

	// Read file
	jsonData, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read save file %s: %w", fullPath, err)
	}

	// Try to parse JSON
	var saveData GameSaveData
	if err := json.Unmarshal(jsonData, &saveData); err != nil {
		return fmt.Errorf("invalid JSON format in save file %s: %w", fullPath, err)
	}

	// Validate basic structure
	if len(saveData.Players) == 0 {
		return fmt.Errorf("save file %s contains no players", fullPath)
	}

	if saveData.GameRules.Name == "" {
		return fmt.Errorf("save file %s has no game rules", fullPath)
	}

	return nil
}

// GetSaveDir returns the save directory path.
func (sm *SaveManager) GetSaveDir() string {
	return sm.SaveDir
}

// Helper methods

// sanitizeFilename removes or replaces invalid characters from a filename.
func (sm *SaveManager) sanitizeFilename(filename string) string {
	// Remove or replace invalid characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	// Remove leading/trailing spaces and dots
	filename = strings.TrimSpace(filename)
	filename = strings.Trim(filename, ".")

	// Ensure filename is not empty
	if filename == "" {
		filename = "save"
	}

	return filename
}

// readGameMetadata reads basic game metadata from a save file without fully loading it.
func (sm *SaveManager) readGameMetadata(fullPath string) (*GameMetadata, error) {
	// Read file
	jsonData, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var saveData GameSaveData
	if err := json.Unmarshal(jsonData, &saveData); err != nil {
		return nil, err
	}

	return &saveData.GameMetadata, nil
}

// Convenience functions for common operations

// SaveGameToFile is a convenience function that creates a SaveManager and saves a game.
func SaveGameToFile(game *Game, saveDir, filename string) error {
	sm, err := NewSaveManager(saveDir)
	if err != nil {
		return err
	}
	return sm.SaveGame(game, filename)
}

// LoadGameFromFile is a convenience function that creates a SaveManager and loads a game.
func LoadGameFromFile(saveDir, filename string) (*Game, error) {
	sm, err := NewSaveManager(saveDir)
	if err != nil {
		return nil, err
	}
	return sm.LoadGame(filename)
}

// ListSaveFiles is a convenience function that creates a SaveManager and lists save files.
func ListSaveFiles(saveDir string) ([]SaveFileInfo, error) {
	sm, err := NewSaveManager(saveDir)
	if err != nil {
		return nil, err
	}
	return sm.ListSaves()
}

// DeleteSaveFile is a convenience function that creates a SaveManager and deletes a save file.
func DeleteSaveFile(saveDir, filename string) error {
	sm, err := NewSaveManager(saveDir)
	if err != nil {
		return err
	}
	return sm.DeleteSave(filename)
}

// ValidateSaveFile is a convenience function that creates a SaveManager and validates a save file.
func ValidateSaveFile(saveDir, filename string) error {
	sm, err := NewSaveManager(saveDir)
	if err != nil {
		return err
	}
	return sm.ValidateSaveFile(filename)
}
