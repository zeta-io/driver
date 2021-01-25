package ginx

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zeta-io/zeta"
	"reflect"
	"strings"
)

const ContextKey = "gin#context#key"

var(
	contextType = "context.Context"
	ginContextType = "gin.Context"
	sourceTags = []string{"query", "body", "path", "header", "cookie", "file"}
)

type Driver struct {
	e *gin.Engine

	serial zeta.Serial
	validator zeta.Validator
	disableValidator bool
	r func(c *gin.Context, data interface{}, err error)
}

func defaultResponse(c *gin.Context, data interface{}, err error){
	if err != nil{
		c.JSON(500, err.Error())
	}
	c.JSON(200, data)
}

func New(e *gin.Engine) *Driver{
	return &Driver{e: e, serial: zeta.DefaultSerial(), validator: zeta.DefaultValidator(), r: defaultResponse}
}

func (d *Driver) Serial(s zeta.Serial) *Driver{
	d.serial = s
	return d
}

func (d *Driver) Validator(v zeta.Validator) *Driver{
	d.validator = v
	return d
}

func (d *Driver) DisableValidator(flag bool) *Driver{
	d.disableValidator = flag
	return d
}

func (d *Driver) Response(r func(c *gin.Context, data interface{}, err error)) *Driver{
	d.r = r
	return d
}

func (d *Driver) Run(addr... string) error{
	return d.e.Run(addr...)
}

func (d *Driver) Option(z *zeta.Zeta) {
	for _, m := range z.Middleware(){
		d.e.Use(d.handleFunc(m))
	}
	for _, m := range z.Mappings(){
		d.Handle(m)
	}
}

func (d *Driver) Handle(mapping zeta.Mapping){
	handleFunc := make([]gin.HandlerFunc, 0)
	for _, m := range mapping.Middleware(){
		handleFunc = append(handleFunc, d.handleFunc(m))
	}
	if mapping.Method() == zeta.MethodAny{
		d.e.Any(mapping.Url(), handleFunc...)
	}else{
		d.e.Handle(string(mapping.Method()), mapping.Url(), handleFunc...)
	}
}

func (d *Driver) handleFunc(call zeta.HandlerFunc) gin.HandlerFunc{
	return func(c *gin.Context){
		ctx := context.WithValue(context.Background(), ContextKey, c)
		if call == nil{
			panic("handler func args is nil.")
		}
		if reflect.TypeOf(call).Kind() != reflect.Func{
			fmt.Println(reflect.TypeOf(call).Kind())
			panic("handler func type must be func.")
		}
		if c.IsAborted(){
			return
		}
		rets, err := d.process(ctx, c, call)
		if err != nil{
			d.r(c, err.Error(), err)
			return
		}
		if len(rets) > 0{
			var data interface{}
			var err error
			for _, ret := range rets{
				if e, ok := ret.Interface().(error); ok{
					err = e
				}else if data == nil{
					data = ret.Interface()
				}
			}
			d.r(c, data, err)
		}
	}
}

func (d *Driver) process(ctx context.Context, c *gin.Context, call interface{}) ([]reflect.Value, error){
	processor, err := newRequestParamsProcessor(c, d.serial)
	if err != nil{
		return nil, err
	}
	typ := reflect.TypeOf(call)
	args := make([]reflect.Value, 0)
	for i := 0; i < typ.NumIn(); i ++{
		in := typ.In(i)
		ptr := false
		if in.Kind() == reflect.Ptr{
			ptr = true
			// handle as element type.
			in = in.Elem()
		}

		var target reflect.Value
		switch in.String() {
		case contextType:
			target = reflect.ValueOf(ctx)
			if ptr{
				target = reflect.ValueOf(&ctx)
			}
		case ginContextType:
			target = reflect.ValueOf(*c)
			if ptr{
				target = reflect.ValueOf(c)
			}
		default:
			if in.Kind() != reflect.Struct{
				continue
			}
			target, err = processRequestParams(processor, in, ptr)
			if err != nil{
				return nil, err
			}
			if ! d.disableValidator{
				err = d.validator.Validate(target.Interface())
				if err != nil{
					return nil, err
				}
			}
		}
		args = append(args, target)
	}
	return reflect.ValueOf(call).Call(args), nil
}

func processRequestParams(processor *requestParamsProcessor, in reflect.Type, ptr bool) (reflect.Value, error){
	obj := reflect.New(in).Elem()
	for i := 0; i < in.NumField(); i ++{
		f := in.Field(i)
		ft := f.Type
		ptr := false
		if ft.Kind() == reflect.Ptr{
			ptr = true
			ft = ft.Elem()
		}

		source, name, defaultValue := parseTag(f.Tag)
		if source == ""{
			continue
		}
		ret, ok, err := processor.process(ft, source, name, defaultValue)
		if err != nil{
			return obj, err
		}
		if ( ! ok || ret == nil) && ptr{
			target := reflect.New(f.Type).Elem()
			obj.FieldByName(f.Name).Set(target)
		}else{
			target := reflect.New(ft).Elem()
			if ret != nil{
				target.Set(reflect.ValueOf(ret))
			}
			if ptr{
				target = target.Addr()
			}
			obj.FieldByName(f.Name).Set(target)
		}
	}
	if ptr{
		obj = obj.Addr()
	}
	return obj, nil
}

func parseTag(tag reflect.StructTag) (source, name string, defaultValue *string){
	if tag == ""{
		return
	}
	for _, tagName := range sourceTags{
		if value, ok := tag.Lookup(tagName); ok{
			name, defaultValue = parseTagValue(value)
			source = tagName
			break
		}
	}
	return
}

func parseTagValue(value string) (name string, defaultValue *string){
	strs := strings.Split(value, ",")
	l := len(strs)
	if l > 0{
		name = strs[0]
	}
	if l > 1{
		defaultValue = &strs[1]
	}
	return
}


