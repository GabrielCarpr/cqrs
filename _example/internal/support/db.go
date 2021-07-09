package support

import (
	"example/internal/errs"
	"log"
	"reflect"
	"regexp"

	"github.com/jackc/pgconn"
	"github.com/lib/pq"
)

func checkSql(text string) string {
	re := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	if re.MatchString(text) {
		return text
	}
	log.Panicf("Stopped SQL injection: %s", text)
	return ""
}

func TransduceError(err error) error {
	switch v := err.(type) {
	case *pq.Error:
		return transducePqError(v)
	case *pgconn.PgError:
		return transduceGormError(v)
	case nil:
		return nil
	default:
		log.Printf("Unknown DB error type: %s", reflect.TypeOf(v))
		return v
	}
}

// Takes a DB error and returns an internal useable error
func transducePqError(err *pq.Error) error {
	switch err.Code {
	case pq.ErrorCode("23505"):
		return errs.UniqueEntityExists

	default:
		log.Printf("Tried to transduce unexpected PQ error: [%s] %s", err.Code, err.Error())
		return err
	}
}

func transduceGormError(err *pgconn.PgError) error {
	switch err.SQLState() {
	case "23505":
		return errs.UniqueEntityExists

	default:
		log.Printf("Tried to transduce unexpected PG error: [%s] %s", err.SQLState(), err.Error())
		return err
	}
}
