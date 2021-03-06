// Package wit implements NLP objects using Wit as the datasource.
package wit

import (
	"errors"
	"fmt"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/aziule/conversation-management/core/nlp"
	"github.com/aziule/conversation-management/core/utils"
	log "github.com/sirupsen/logrus"
)

var (
	// @todo: move these variables to the core and return only core errors
	// related to the Parser interface itself
	ErrCouldNotParseJson       = errors.New("Could not parse JSON")
	ErrCouldNotParseJsonObject = errors.New("Could not parse object from JSON")
	ErrMissingKey              = func(key string) error { return errors.New(fmt.Sprintf("Missing key: %s", key)) }
	ErrCouldNotCastValue       = func(key, expectedType string) error {
		return errors.New(fmt.Sprintf("Could not cast %s to %s", expectedType))
	}
	ErrUnhandledDataType = func(dataType string) error { return errors.New(fmt.Sprintf("Unhandled data type %s", dataType)) }

	// DefaultDataTypeMap is the default data type map to be used with Wit.
	// For now, this is highly coupled with Wit's data types and should
	// be updated every time a change is made to Wit.
	// It is initialised in the init() function
	defaultDataTypeMap nlp.DataTypeMap
)

func init() {
	defaultDataTypeMap = make(nlp.DataTypeMap)
	defaultDataTypeMap["nb_persons"] = nlp.IntEntity
	defaultDataTypeMap["intent"] = nlp.IntentEntity
	defaultDataTypeMap["datetime"] = nlp.DateTimeEntity

	nlp.RegisterParserBuilder("wit", newParser)
}

// witParser is the NLP parser for Wit.
// It implements the nlp.Parser interface.
type witParser struct {
	dataTypeMap nlp.DataTypeMap
}

// newParser is the constructor method for witParser
func newParser(conf utils.BuilderConf) (interface{}, error) {
	return &witParser{
		dataTypeMap: defaultDataTypeMap,
	}, nil
}

// ParseNlpData parses raw data and returns parsed data
func (parser *witParser) ParseNlpData(rawData []byte) (*nlp.ParsedData, error) {
	var intent *nlp.ParsedIntent
	var entities []*nlp.ParsedEntity

	data, err := jason.NewObjectFromBytes(rawData)

	if err != nil {
		log.WithField("rawData", string(rawData)).Infof("Could not parse JSON: %s", err)
		return nil, ErrCouldNotParseJson
	}

	for key, value := range data.Map() {
		dataType, ok := parser.dataTypeMap[key]

		if !ok {
			log.WithField("key", key).Warnf("Data type is not handled: %s", key)
			continue
		}

		switch dataType {
		case nlp.IntentEntity:
			i, err := toIntent(value)

			if err != nil {
				log.WithField("dataType", dataType).Warnf("Could not convert value to DataTypeIntent: %s", err)
				continue
			}

			intent = i
			break
		default:
			entity, err := toEntity(value, key, dataType)

			if err != nil {
				log.WithFields(log.Fields{
					"dataType": dataType,
					"key":      key,
				}).Warnf("Could not convert value to entity: %s", err)
				continue
			}

			entities = append(entities, entity)
			break
		}
	}

	return nlp.NewParsedData(intent, entities), nil
}

// toIntent converts a jason intent to a built-in NLP representation of an intent
// Returns an error if the JSON is malformed
func toIntent(value *jason.Value) (*nlp.ParsedIntent, error) {
	object, err := value.ObjectArray()

	if err != nil {
		return nil, ErrCouldNotParseJsonObject
	}

	// Handle single intents only
	intentName, err := object[0].GetString("value")

	if err != nil {
		return nil, ErrMissingKey("value")
	}

	return nlp.NewParsedIntent(intentName), nil
}

// toEntity converts a jason entity to a built-in NLP representation of an entity
// Returns an error if the JSON is malformed or if we do not handle the data type correctly
func toEntity(value *jason.Value, name string, dataType nlp.EntityType) (*nlp.ParsedEntity, error) {
	object, err := value.ObjectArray()

	if err != nil {
		return nil, ErrCouldNotParseJsonObject
	}

	for _, e := range object {
		confidence, err := e.GetFloat64("confidence")

		if err != nil {
			return nil, ErrCouldNotCastValue("confidence", "float64")
		}

		switch dataType {
		case nlp.IntEntity:
			value, err := e.GetInt64("value")

			if err != nil {
				return nil, ErrCouldNotCastValue("value", "int64")
			}

			// @todo: handle roles
			return nlp.NewParsedIntEntity(name, float32(confidence), int(value), ""), nil
		case nlp.DateTimeEntity:
			_, err := e.GetString("value")

			if err != nil {
				// If there's an error, then look for interval datetimes, parsed as "from" & "to"
				from, err := e.GetObject("from")

				if err != nil {
					return nil, ErrMissingKey("from")
				}

				fromTime, fromGranularity, err := extractDateTimeInformation(from)

				if err != nil {
					return nil, err
				}

				to, err := e.GetObject("to")

				if err != nil {
					return nil, ErrMissingKey("to")
				}

				toTime, toGranularity, err := extractDateTimeInformation(to)

				if err != nil {
					return nil, err
				}

				// @todo: handle roles
				return nlp.NewParsedDateTimeIntervalEntity(
					name,
					float32(confidence),
					fromTime,
					toTime,
					fromGranularity,
					toGranularity,
					"",
				), nil
			}

			t, granularity, err := extractDateTimeInformation(e)

			if err != nil {
				return nil, err
			}

			return nlp.NewParsedSingleDateTimeEntity(name, float32(confidence), t, granularity, ""), nil
		}
	}

	return nil, ErrUnhandledDataType(string(dataType))
}

// extractDateTimeInformation extracts useful date time information from JSON
// This is heavily coupled with what Wit returns us
func extractDateTimeInformation(object *jason.Object) (time.Time, nlp.DateTimeGranularity, error) {
	value, err := object.GetString("value")

	if err != nil {
		return time.Time{}, "", ErrMissingKey("value")
	}

	t, err := utils.ParseTime(value)

	if err != nil {
		return time.Time{}, "", err
	}

	grain, err := object.GetString("grain")

	if err != nil {
		return time.Time{}, "", ErrMissingKey("grain")
	}

	// @todo: use a converter that will check that the granularity exists
	return t, nlp.DateTimeGranularity(grain), nil
}
