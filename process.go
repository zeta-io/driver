package ginx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zeta-io/zeta"
	"io/ioutil"
	"net/url"
	"reflect"
	"strings"
)

var (
	parameterIsNullError = errors.New("parameter is null. ")
)

type Values map[string][]string

func (v Values) Get(key string) (string, bool){
	if vs, ok := v[key]; ok{
		if len(vs) > 0{
			return vs[0], true
		}
	}
	return "", false
}

func (v Values) GetArray(key string) ([]string, bool){
	vs, ok := v[key]
	return vs, ok
}

type requestParamsProcessor struct {
	c *gin.Context
	serial zeta.Serial

	body string
	forms Values
	queries Values
	contentType string
}

func newRequestParamsProcessor(c *gin.Context, serial zeta.Serial) (*requestParamsProcessor, error){
	contentType := contentType(c)
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil{
		return nil, err
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	queries := Values{}
	err = parseQuery(queries, c.Request.URL.RawQuery)
	if err != nil{
		return nil, err
	}

	forms := Values{}
	if contentType == string(zeta.ContentTypePostForm){
		err = parseQuery(forms, string(body))
		if err != nil{
			return nil, err
		}
	}else if contentType == string(zeta.ContentTypeFormData){
		//TODO parse multipart/form-data
	}

	return &requestParamsProcessor{
		c: c,
		serial: serial,
		contentType: contentType,
		body: string(body),
		queries: queries,
		forms: forms,
	}, nil
}

func contentType(c *gin.Context) string{
	return strings.TrimSpace(strings.Split(c.ContentType(), ";")[0])
}

func (p *requestParamsProcessor) process(t reflect.Type, source, name string, defaultValue *string) (ret interface{}, ok bool, err error){
	switch source {
	case "query":
		ret, ok, err = p.processQuery(t, name)
	case "body":
		ret, ok, err = p.processBody(t, name)
	case "path":
		ret, ok, err = p.processPath(t, name)
	case "header":
		ret, ok, err = p.processHeader(t, name)
	case "file":
		ret, ok, err = p.processFile(t, name)
	case "cookie":
		ret, ok, err = p.processCookie(t, name)
	default:
		ret, ok, err = nil, false, errors.New(fmt.Sprintf("unsupport params source: %v", source))
		return
	}
	if ! ok && defaultValue != nil{
		ret, err = p.serial.DeSerial(*defaultValue, t)
		ok = true
	}
	return
}


func (p *requestParamsProcessor) processQuery(t reflect.Type, name string) (interface{}, bool, error){
	src := interface{}(nil)
	ok := false
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice{
		src, ok = p.queries.GetArray(name)
	}else{
		src, ok = p.queries.Get(name)
	}
	if ! ok{
		return nil, false, nil
	}
	v, err := p.serial.DeSerial(src, t)
	return v, true, err
}

func (p *requestParamsProcessor) processPath(t reflect.Type, name string) (interface{}, bool, error){
	value, ok := p.c.Params.Get(name)
	if ! ok {
		return nil, false, nil
	}
	v, err := p.serial.DeSerial(value, t)
	return v, true, err
}

func (p *requestParamsProcessor) processFile(t reflect.Type, name string) (interface{}, bool, error){
	file, err := p.c.FormFile(name)
	return file, err == nil, err
}

func (p *requestParamsProcessor) processHeader(t reflect.Type, name string) (interface{}, bool, error){
	header := p.c.GetHeader(name)
	if header == ""{
		return nil, false, nil
	}
	v, err := p.serial.DeSerial(header, t)
	return v, true, err
}

func (p *requestParamsProcessor) processCookie(t reflect.Type, name string) (interface{}, bool, error){
	value, err := p.c.Cookie(name)
	if err != nil {
		return nil, false, err
	}
	v, err := p.serial.DeSerial(value, t)
	return v, true, err
}

func (p *requestParamsProcessor) processFormData(t reflect.Type, name string) (interface{}, bool, error){
	src := interface{}(nil)
	ok := false
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice{
		src, ok = p.forms.GetArray(name)
	}else{
		src, ok = p.forms.Get(name)
	}
	v, err := p.serial.DeSerial(src, t)
	return v, ok, err
}

func (p *requestParamsProcessor) processBody(t reflect.Type, name string) (interface{}, bool, error){
	contentType := contentType(p.c)
	if contentType == string(zeta.ContentTypeJSON) {
		if name != ""{

		}else{
			v, err := p.serial.DeSerial(p.body, t)
			return v, true, err
		}
	}else if contentType == string(zeta.ContentTypePostForm){
		if name != ""{
			return p.processFormData(t, name)
		}else{

		}
	}
	return nil, true, nil
}

func parseQuery(values Values, query string) error {
	var err error
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := queryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		for _, sub := range strings.Split(value, ","){
			value, err1 = queryUnescape(sub)
			if err1 != nil {
				if err == nil {
					err = err1
				}
				continue
			}
			values[key] = append(values[key], value)
		}
	}
	return err
}

func queryUnescape(v string) (string, error){
	return url.QueryUnescape(v)
}