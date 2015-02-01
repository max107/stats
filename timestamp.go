package main

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	tt := time.Time(t).UTC()
	if y := tt.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Timestamp: year outside of range [0,9999]")
	}
	if y := tt.Year(); y == 1 {
		return []byte{}, nil
	}
	return []byte(tt.Format(`"` + time.RFC3339Nano + `"`)), nil
}

func (t Timestamp) GetBSON() (interface{}, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t), nil
}

func (t *Timestamp) SetBSON(raw bson.Raw) error {
	var tm time.Time

	if err := raw.Unmarshal(&tm); err != nil {
		return err
	}

	*t = Timestamp(tm)
	return nil
}

func (t Timestamp) String() string {
	return time.Time(t).UTC().String()
}
