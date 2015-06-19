package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Conf map[string]interface{}
}

func NewConfig() *Config {
	return &Config{Conf: make(map[string]interface{})}
}

func NewError() *Config {
	return fmt.Errorf("Connect-Server Error: the %s is not found in config file", key)
}

func (config *Config) ParseFile(file *os.File) error {
	readFile := bufio.NewReader(file)
	for {
		line, _, err := readFile.ReadLine()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			err = fmt.Errorf("Connect-Server ParseFile Error: Read the file error: %v", err)
			return err
		}

		str := strings.TrimSpace(string(line))
		if len(str) == 0 || str[0] == '#' {
			continue
		}

		s := strings.Split(str, "=")
		if strings.TrimSpace(s[1])[0] == '{' {
			value := make([]interface{}, 0)
			for {
				line, _, err := readFile.ReadLine()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					err = fmt.Errorf("Connect-Server ParseFile Error: Read the file error: %v", err)
					return err
				}
				str := strings.TrimSpace(string(line))
				if len(str) == 0 || str[0] == '#' {
					continue
				}
				if str[0] == '}' {
					break
				}
				value = append(value, str)
			}
			config.Conf[strings.TrimSpace(s[0])] = value
		} else {
			config.Conf[strings.TrimSpace(s[0])] = strings.TrimSpace(s[1])
		}
	}
	return nil
}

func (config *Config) Read(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		err = fmt.Errorf("Config Read Error: Open %s error %v", fileName, err)
		return err
	}
	defer file.Close()

	err = config.ParseFile(file)
	if err != nil {
		return err
	}
	return nil
}

func (config *Config) Get(key string) (interface{}, error) {
	if value, ok := config.Conf[key]; ok {
		return value, nil
	}

	err := NewError(key)

	return nil, err
}

func (config *Config) GetString(key string) (value string, err error) {
	if val, ok := config.Conf[key]; ok {
		if value, ok = val.(string); ok {
			return value, nil
		}
	}

	err = NewError(key)
	return value, err
}
func (config *Config) GetStringArray(key string) (value []string, err error) {
	if val, ok := config.Conf[key]; ok {
		for _, v := range val.([]interface{}) {
			if v1, ok := v.(string); ok {
				value = append(value, v1)
			}
		}
		return value, nil
	}
	err = NewError(key)

	return value, err
}

func (config *Config) GetInt64(key string) (value int64, err error) {
	if val, ok := config.Conf[key]; ok {
		if v, ok := val.(string); ok {
			value, err = strconv.ParseInt(v, 0, 64)
			if err != nil {
				return value, err
			}
			return value, nil
		}
	}

	err = NewError(key)
	return value, err
}

func (config *Config) GetInt32(key string) (int32, error) {
	if val, ok := config.Conf[key]; ok {
		if v, ok := val.(string); ok {
			value, err := strconv.ParseInt(v, 0, 32)
			if err != nil {
				return int32(value), err
			}
			return int32(value), nil
		}
	}
	err := NewError(key)
	return 0, err
}

func (config *Config) GetInt(key string) (value int, err error) {
	if val, ok := config.Conf[key]; ok {
		if v, ok := val.(string); ok {
			value, err = strconv.Atoi(v)
			if err != nil {
				return value, err
			}

			return value, nil
		}
	}

	err = NewError(key)
	return value, err
}

func (config *Config) Set(key string, value interface{}) bool {
	config.Conf[key] = value
	return true
}
