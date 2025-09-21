package models

import (
	"encoding/json"
	"time"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Application string    `json:"application"`
	Level       LogLevel  `json:"level"`
	Message     string    `json:"message"`
}

func (l *LogEntry) ToJson() ([]byte, error) {
	return json.Marshal(l)
}

func FromJson(data []byte) (*LogEntry, error) {
	var entry LogEntry
	err := json.Unmarshal(data, &entry)
	return &entry, err
}
