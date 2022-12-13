package gvalue

// DBWInstance db web中间件写实例
//const DBWInstance =

// DBRInstance db web中间件读实例
//const DBRInstance = "__db_read_instance"

type DBContext struct {
	Key string
}

var (
	DBWriteInstance = DBContext{Key: "__db_write_instance"}
	DBReadInstance  = DBContext{Key: "__db_read_instance"}
)

type baggageKey int
const DBBaggageKey baggageKey = iota

func (d DBContext) String() string {
	return d.Key
}

type Baggage struct {
	TraceID string
	IP      string
}
