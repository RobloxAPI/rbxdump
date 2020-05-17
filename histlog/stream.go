package histlog

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"time"
)

type Stream []Token

func (s Stream) MarshalJSON() (b []byte, err error) {
	b = append(b, []byte(`[`)...)
	for i, token := range s {
		if i > 0 {
			b = append(b, ',')
		}
		bsub, err := token.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b = append(b, bsub...)
	}
	b = append(b, ']')
	return b, nil
}

func (s *Stream) UnmarshalJSON(b []byte) error {
	type jStream []struct {
		Type    string
		Action  string
		Build   string
		GUID    string
		Time    time.Time
		Version Version
		Value   string
	}
	var stream jStream
	if err := json.Unmarshal(b, &stream); err != nil {
		return err
	}
	for _, token := range stream {
		switch token.Type {
		case "Job":
			*s = append(*s, &Job{
				Action:  token.Action,
				Build:   token.Build,
				GUID:    token.GUID,
				Time:    token.Time,
				Version: token.Version,
			})
		case "Status":
			t := Status(token.Value)
			*s = append(*s, &t)
		case "Raw":
			t := Raw(token.Value)
			*s = append(*s, &t)
		}
	}
	return nil
}

type Token interface {
	token()
	json.Marshaler
	json.Unmarshaler
}

type Job struct {
	Action  string
	Build   string
	GUID    string
	Time    time.Time
	Version Version
}

func (Job) token() {}

func (j *Job) MarshalJSON() (b []byte, err error) {
	var buf bytes.Buffer
	var c []byte
	buf.WriteString(`{"Type":"Job","Action":`)
	c, _ = json.Marshal(j.Action)
	buf.Write(c)
	buf.WriteString(`,"Build":`)
	c, _ = json.Marshal(j.Build)
	buf.Write(c)
	buf.WriteString(`,"GUID":`)
	c, _ = json.Marshal(j.GUID)
	buf.Write(c)
	buf.WriteString(`,"Time":`)
	c, _ = j.Time.MarshalJSON()
	buf.Write(c)
	if !j.Version.Empty() {
		buf.WriteString(`,"Version":`)
		c, _ = j.Version.MarshalJSON()
		buf.Write(c)
	}
	buf.WriteString(`}`)
	return buf.Bytes(), nil
}

func (j *Job) UnmarshalJSON(b []byte) error {
	type jJob struct {
		Type    string
		Action  string
		Build   string
		GUID    string
		Time    time.Time
		Version Version
	}
	job := jJob{}
	if err := json.Unmarshal(b, &job); err != nil {
		return err
	}
	if job.Type != "Job" {
		return nil
	}
	j.Action = job.Action
	j.Build = job.Build
	j.GUID = job.GUID
	j.Time = job.Time
	j.Version = job.Version
	return nil
}

type Version struct {
	Major, Minor, Maint, Build int
}

var versionGrammar = regexp.MustCompile(`` +
	`^(\d+)\.(\d+)\.(\d+)\.(\d+)$` +
	`|^(\d+), (\d+), (\d+), (\d+)$`,
)

func VersionFromString(s string) (v Version, ok bool) {
	r := versionGrammar.FindStringSubmatch(s)
	if r[0] != "" {
		if r[1] != "" {
			v.Major, _ = strconv.Atoi(r[1])
			v.Minor, _ = strconv.Atoi(r[2])
			v.Maint, _ = strconv.Atoi(r[3])
			v.Build, _ = strconv.Atoi(r[4])
			return v, true
		} else if r[5] != "" {
			v.Major, _ = strconv.Atoi(r[5])
			v.Minor, _ = strconv.Atoi(r[6])
			v.Maint, _ = strconv.Atoi(r[7])
			v.Build, _ = strconv.Atoi(r[8])
			return v, true
		}
	}
	return v, false
}

func (v Version) Empty() bool {
	return v.Major == 0 && v.Minor == 0 && v.Maint == 0 && v.Build == 0
}

func (a Version) Compare(b Version) int {
	switch {
	case a.Major < b.Major:
		return -1
	case a.Major > b.Major:
		return 1
	case a.Minor < b.Minor:
		return -1
	case a.Minor > b.Minor:
		return 1
	case a.Maint < b.Maint:
		return -1
	case a.Maint > b.Maint:
		return 1
	case a.Build < b.Build:
		return -1
	case a.Build > b.Build:
		return 1
	}
	return 0
}

func (v Version) MarshalJSON() (b []byte, err error) {
	b = append(b, '"')
	b = strconv.AppendUint(b, uint64(v.Major), 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, uint64(v.Minor), 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, uint64(v.Maint), 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, uint64(v.Build), 10)
	b = append(b, '"')
	return b, nil
}

func (v *Version) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*v, _ = VersionFromString(s)
	return nil
}

func (v Version) String() string {
	return strconv.Itoa(v.Major) +
		"." + strconv.Itoa(v.Minor) +
		"." + strconv.Itoa(v.Maint) +
		"." + strconv.Itoa(v.Build)
}

type Status string

func (Status) token() {}

func (s Status) MarshalJSON() (b []byte, err error) {
	var buf bytes.Buffer
	var c []byte
	buf.WriteString(`{"Type":"Status","Value":`)
	c, _ = json.Marshal(string(s))
	buf.Write(c)
	buf.WriteString(`}`)
	return buf.Bytes(), nil
}

func (s *Status) UnmarshalJSON(b []byte) error {
	type jStatus struct {
		Type  string
		Value string
	}
	status := jStatus{}
	if err := json.Unmarshal(b, &status); err != nil {
		return err
	}
	if status.Type != "Status" {
		return nil
	}
	*s = Status(status.Value)
	return nil
}

type Raw string

func (Raw) token() {}

func (r Raw) MarshalJSON() (b []byte, err error) {
	var buf bytes.Buffer
	var c []byte
	buf.WriteString(`{"Type":"Raw","Value":`)
	c, _ = json.Marshal(string(r))
	buf.Write(c)
	buf.WriteString(`}`)
	return buf.Bytes(), nil
}

func (r *Raw) UnmarshalJSON(b []byte) error {
	type jRaw struct {
		Type  string
		Value string
	}
	raw := jRaw{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if raw.Type != "Raw" {
		return nil
	}
	*r = Raw(raw.Value)
	return nil
}
