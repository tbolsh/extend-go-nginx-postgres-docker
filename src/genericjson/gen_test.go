package genericjson

import (
	"encoding/json"
	"log"
	"testing"
)

var raw = `{"test_string":"str", "test_int":1, "test_float":1.0, "test_bool":true,
    "test_arr":["str", 1, 1.0, true]}`
var ui = `
  {
    "fhir_patient": {
      "entry": [
        {
          "resource": {
            "address": [
              {
                "city": "Seattle",
                "country": "United States of America",
                "line": [
                  "123 Any Street"
                ],
                "postalCode": "98124",
                "state": "Washington",
                "text": "123 Any Street",
                "type": "both",
                "use": "work"
              }
            ],
            "birthDate": "1980-01-01",
            "gender": "female",
            "identifier": [
              {
                "system": "org.providence.dig.epic.provorca",
                "type": {
                  "coding": [
                    {
                      "code": "MR",
                      "display": "Medical record number",
                      "system": "http://hl7.org/fhir/v2/0203"
                    }
                  ]
                },
                "use": "official",
                "value": "E20008100372"
              },
              {
                "system": "org.providence.dig.epic.provorca.mychart",
                "use": "official",
                "value": "TESTODHPPARENT"
              },
              {
                "system": "http://hl7.org/fhir/sid/us-ssn",
                "type": {
                  "coding": [
                    {
                      "code": "SB",
                      "display": "Social Beneficiary Identifier",
                      "system": "http://hl7.org/fhir/identifier-type"
                    }
                  ]
                },
                "use": "official",
                "value": "xxx-xx-1352"
              }
            ],
            "maritalStatus": {
              "coding": [
                {
                  "code": "M",
                  "display": "Married",
                  "system": "http://hl7.org/fhir/v3/MaritalStatus"
                }
              ]
            },
            "name": [
              {
                "family": "Zzztrash",
                "given": [
                  "Odhpparent"
                ],
                "text": "Zzztrash,Odhpparent",
                "use": "official"
              }
            ],
            "resourceType": "Patient",
            "telecom": [
              {
                "system": "email",
                "use": "work",
                "value": "ecare@mailinator.com"
              },
              {
                "system": "email",
                "use": "work",
                "value": "ron.ewald@providence.org"
              },
              {
                "system": "email",
                "use": "work",
                "value": "nataliya.slyusarchuk@providence.org"
              },
              {
                "system": "email",
                "use": "work",
                "value": "jonathan.becker@providence.org"
              },
              {
                "system": "phone",
                "use": "home",
                "value": "222-222-2222"
              }
            ]
          }
        }
      ],
      "resourceType": "Bundle",
      "total": 1,
      "type": "searchset"
    },
    "fhir_person": {
      "entry": [
        {
          "resource": {
            "id": "d30817b5-7632-444a-a010-6a3dfd014a58",
            "identifier": [
              {
                "system": "org.psjh.providence.patients",
                "value": "5b77117caf1e9a7b604dac04"
              }
            ],
            "link": [
              {
                "target": {
                  "identifier": {
                    "system": "org.providence.dig.epic.provorca",
                    "value": "E20008100372"
                  },
                  "reference": "Patient"
                }
              }
            ],
            "resourceType": "Person"
          }
        }
      ],
      "resourceType": "Bundle",
      "total": 1,
      "type": "searchset"
    }
  }`

func TestSet(t *testing.T) {
	var g GenJson
	err := json.Unmarshal([]byte(ui), &g)
	if err != nil {
		t.Error("error unmarshaling user_info")
	}
	err = g.Set("Oppa!", "fhir_person", "entry", 0, "resource", "link", 0, "target", "identifier", "value")
	if err != nil {
		t.Errorf("error setting value %v", err)
	}
	v, err := g.String("fhir_person", "entry", 0, "resource", "link", 0, "target", "identifier", "value")
	if err != nil {
		t.Errorf("error reading value back %v", err)
	}
	if v != "Oppa!" {
		t.Errorf("Value was not set, should be 'Oppa!', but is '%s'", v)
	}
}

func TestUI(t *testing.T) {
	var g GenJson
	err := json.Unmarshal([]byte(ui), &g)
	if err != nil {
		t.Error("error unmarshaling user_info")
	}
	second := func(i interface{}, err error) error { return err }
	first := func(i GenJson, err error) interface{} { return i.Any }
	_, p, ok := g.ScanObject(func(obj interface{}) bool {
		s := FromGeneric(obj)
		retval := first(s.Unwind("type")) == nil &&
			second(s.Unwind("type")) == nil &&
			second(s.Unwind("value")) == nil &&
			second(s.Unwind("system")) == nil
		return retval
	}, "fhir_patient", "entry", -1, "resource", "identifier")
	if !ok {
		t.Error("Not found EHR entry in fhir_patient")
	}
	log.Printf("path length is %d %v", len(p), p)
	_, _, ok = g.ScanObject(func(obj interface{}) bool {
		s := FromGeneric(obj)
		t, e1 := s.String("system")
		e, e2 := s.String("value")
		retval := e1 == nil && e2 == nil && t == "email" && e != ""
		log.Printf("Predicate2 %v => %v", obj, retval)
		return retval
	}, append(p, "telecom", -1)...)
	//if !ok { t.Error("Not found EHR entry email in fhir_patient") }
}

func TestRaw(t *testing.T) {
	var g GenJson
	err := json.Unmarshal([]byte(raw), &g)
	if err != nil {
		t.Errorf("Error unmarshalling test Json: '%v'", err)
	}
	if v, err := g.String("test_string"); err != nil || "str" != v {
		t.Errorf("Error getting string err: '%v', val: '%s'", err, v)
	}
	if v, err := g.String("test_arr", 0); err != nil || "str" != v {
		t.Errorf("Error getting string from array err: '%v', val: '%s'", err, v)
	}
	if v, err := g.Int("test_int"); err != nil || 1 != v {
		t.Errorf("Error getting int err: '%v', val: '%d'", err, v)
	}
	if v, err := g.Int("test_arr", 1); err != nil || 1 != v {
		t.Errorf("Error getting int from array err: '%v', val: '%d'", err, v)
	}
	if v, err := g.Float("test_float"); err != nil || v < 0.99 || v > 1.01 {
		t.Errorf("Error getting float err: '%v', val: '%v'", err, v)
	}
	if v, err := g.Float("test_arr", 2); err != nil || v < 0.99 || v > 1.01 {
		t.Errorf("Error getting float from array err: '%v', val: '%v'", err, v)
	}
	if v, err := g.Bool("test_bool"); err != nil || !v {
		t.Errorf("Error getting bool err: '%v', val: '%v'", err, v)
	}
	if v, err := g.Bool("test_arr", 3); err != nil || !v {
		t.Errorf("Error getting bool from array err: '%v', val: '%v'", err, v)
	}
}
