package gwf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const defaultMaxMemory = 32 << 20 // 32M
type Context struct {
	W             http.ResponseWriter
	R             *http.Request
	engine        *Engine
	postFormCache url.Values
}

func (c *Context) initPostForm() {
	err := c.R.ParseMultipartForm(defaultMaxMemory)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		log.Fatal(err)
	}

	c.postFormCache = c.R.PostForm
}

func (c *Context) HTML(code int, html string) error {
	c.W.WriteHeader(code)
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := c.W.Write([]byte(html))
	return err
}

func (c *Context) JSON(code int, data any) error {
	c.W.WriteHeader(code)
	c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(jsonData)
	return err
}

func (c *Context) QueryMap(key string) map[string]string {
	query := c.R.URL.Query()

	return getMapFromValues(query, key)
}

func getMapFromValues(values url.Values, key string) map[string]string {
	dicts := make(map[string]string)
	for k, v := range values {
		if i := strings.Index(k, "["); i >= 1 && k[0:i] == key {
			if j := strings.Index(k[i+1:], "]"); j > 0 {
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts
}

func (c *Context) PostForm(key string) string {
	c.initPostForm()
	v := c.postFormCache[key]
	return v[0]
}
func (c *Context) PostFormArray(key string) []string {
	c.initPostForm()
	v := c.postFormCache[key]
	return v
}
func (c *Context) PostFormMap(key string) map[string]string {
	c.initPostForm()
	return getMapFromValues(c.postFormCache, key)
}

func (c *Context) DealJSON(obj any) error {
	if c.R.Body == nil {
		return errors.New("invalid request")
	}
	body, err := io.ReadAll(c.R.Body)
	if err != nil {
		return err
	}
	c.R.Body = io.NopCloser(bytes.NewBuffer(body))
	// err = validateParam(obj, bytes.NewBuffer(body))
	// if err != nil {
	// 	return err
	// }
	decoder := json.NewDecoder(bytes.NewBuffer(body))
	err = decoder.Decode(obj)
	if err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

func validateParam(obj any, body io.Reader) error {
	objV := reflect.ValueOf(obj)
	if objV.Kind() != reflect.Pointer {
		return errors.New("obj need pointer")
	}

	var err error
	elem := objV.Elem().Interface()
	fmt.Println(elem)
	of := reflect.ValueOf(elem)
	fmt.Println(of.Kind())
	switch of.Kind() {
	case reflect.Struct:
		m := make(map[string]interface{})
		decoder := json.NewDecoder(body)
		err = decoder.Decode(&m)
		if err != nil {
			return err
		}
		for i := 0; i < of.NumField(); i++ {
			field := of.Type().Field(i)
			name := field.Name
			jsonName := field.Tag.Get("json")
			if jsonName != "" {
				name = jsonName
			}
			required := field.Tag.Get("bind")
			if _, exist := m[name]; !exist && required == "required" {
				return errors.New(fmt.Sprintf("filed [%s] is not exist", name))
			}
		}
	default:
	}
	return err
}
