package confidential

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

const (
	AlgorithmNacl  = "nacl"
	AlgorithmEcdsa = "ecdsa"
)

type ConfidentialEntry struct {
	Prefix       string `yaml:"prefix"`
	Algorithm    string `yaml:"algorithm"`
	PublicKey    string `yaml:"publickey"`
	Confidential bool   `yaml:"confidential"`
	ReadOnly     bool   `yaml:"readonly"`
}

type SorobanConfig struct {
	Confidential []ConfidentialEntry `yaml:"confidential"`
}

var (
	DefaultSorobanConfig SorobanConfig

	sorobanRegexpMap    map[string]*regexp.Regexp
	sorobanConfigLocker sync.Mutex
)

func init() {
	sorobanRegexpMap = make(map[string]*regexp.Regexp)
}

func (p *SorobanConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, p)
}

func ConfigLoad(filename string) SorobanConfig {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithError(err).WithField("Filename", filename).Error("Failed to load config")
		return SorobanConfig{}
	}
	var config SorobanConfig
	if err := config.Parse(data); err != nil {
		log.WithError(err).WithField("Filename", filename).Error("Failed to parse config")
	}
	return config
}

func ConfigWatcher(ctx context.Context, filename string) {
	if len(filename) == 0 {
		return // Noop
	}
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		log.WithError(err).WithField("Filename", filename).Warning("Config file not found")
	}

	DefaultSorobanConfig = ConfigLoad(filename)

	// configure fs watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).WithField("Filename", filename).Error("Failed to create ConfigWatcher")
		return
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Reloading config file")

					log.Println("modified file:", event.Name)
					DefaultSorobanConfig = ConfigLoad(event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.WithError(err).WithField("Filename", filename).Error("ConfigWatcher error")

			case <-ctx.Done():
				return
			}
		}
	}()

	// Add a path.
	err = watcher.Add(filename)
	if err != nil {
		log.WithError(err).WithField("Filename", filename).Error("Failed to watch config file")
	}

	<-ctx.Done()
}

func wildCardToRegexp(pattern string) string {
	components := strings.Split(pattern, "*")
	if len(components) == 1 {
		// if len is 1, there are no *'s, return exact match pattern
		return "^" + pattern + "$"
	}
	var result strings.Builder
	for i, literal := range components {

		// Replace * with .*
		if i > 0 {
			result.WriteString(".*")
		}

		// Quote any regular expression meta characters in the
		// literal text.
		result.WriteString(regexp.QuoteMeta(literal))
	}
	return "^" + result.String() + "$"
}

func match(pattern string, value string) bool {
	sorobanConfigLocker.Lock()
	defer sorobanConfigLocker.Unlock()

	str := wildCardToRegexp(pattern)
	if _, ok := sorobanRegexpMap[str]; !ok {
		re, err := regexp.Compile(wildCardToRegexp(pattern))
		if err != nil {
			log.WithError(err).WithField("Pattern", pattern).Error("Failed to Compile regexp")
			return false
		}
		sorobanRegexpMap[str] = re
	}
	if re, ok := sorobanRegexpMap[str]; ok {
		return re.MatchString(value)
	}

	return false
}

func GetConfidentialInfo(directory, publicKey string) ConfidentialEntry {
	var entries []ConfidentialEntry

	// find all matching prefix
	for _, entry := range DefaultSorobanConfig.Confidential {
		if match(entry.Prefix, directory) {
			entries = append(entries, entry)
		}
	}

	var result ConfidentialEntry
	// find first matching publicKey if exists
	for _, entry := range entries {
		if entry.PublicKey == publicKey {
			result = entry
		}
	}

	// some matched prefix exists but with no matching publicKey
	if len(result.Prefix) == 0 && len(entries) > 0 {
		result = entries[0] // use first entry
	}
	if len(publicKey) == 0 && len(entries) > 0 {
		return entries[0]
	}
	return result
}
