package ambiguity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var ErrCommandNotFoundInStore = errors.New("command not found in store")

type CommandStore interface {
	Save(command *discordgo.ApplicationCommand) error
	Load(name string) (discordgo.ApplicationCommand, error)
}

type NoOpCommandStore struct{}

func (n *NoOpCommandStore) Save(command *discordgo.ApplicationCommand) error {
	return nil
}

func (n *NoOpCommandStore) Load(name string) (discordgo.ApplicationCommand, error) {
	return discordgo.ApplicationCommand{}, ErrCommandNotFoundInStore
}

type JsonCommandStore struct {
	mu       sync.RWMutex
	filepath string
	commands map[string]discordgo.ApplicationCommand
}

func NewJsonCommandStore(filepath string) (*JsonCommandStore, error) {
	j := &JsonCommandStore{
		filepath: filepath,
		commands: make(map[string]discordgo.ApplicationCommand),
	}

	if err := j.loadFile(); err != nil {
		return nil, err
	}

	return j, nil
}

func (j *JsonCommandStore) Save(command *discordgo.ApplicationCommand) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.commands[command.Name] = *command
	return j.saveFile()
}

func (j *JsonCommandStore) Load(name string) (discordgo.ApplicationCommand, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	command, ok := j.commands[name]
	if !ok {
		return discordgo.ApplicationCommand{}, ErrCommandNotFoundInStore
	}

	return command, nil
}

func (j *JsonCommandStore) saveFile() error {
	content, err := json.MarshalIndent(j.commands, "", "\t")
	if err != nil {
		return fmt.Errorf("JSON marshal: %w", err)
	}

	if err := os.WriteFile(j.filepath, content, 0644); err != nil {
		return fmt.Errorf("write JSON file: %w", err)
	}

	return nil
}

func (j *JsonCommandStore) loadFile() error {
	content, err := os.ReadFile(j.filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if err := json.Unmarshal(content, &j.commands); err != nil {
		return fmt.Errorf("read JSON file: %w", err)
	}

	return nil
}
