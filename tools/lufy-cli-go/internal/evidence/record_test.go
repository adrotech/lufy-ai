package evidence

import (
	"strings"
	"testing"
)

func TestAssessmentDoesNotTurnMissingEvidenceIntoSuccess(t *testing.T) {
	cases := []struct{ body, want string }{
		{`{"checks":[{"name":"CLI regression","result":"passed","evidence":"go test ./...; exit 0"}]}`, "supported"},
		{`{"checks":[{"name":"editorial","result":"passed","evidence":"Revisión del texto"}],"findings":[{"severity":"critical","summary":"Se alteró el significado"}]}`, "failed"},
		{`{"checks":[{"name":"datos","result":"inconclusive","evidence":"La consulta no terminó"}]}`, "inconclusive"},
		{`{"checks":[{"name":"build","result":"not_applicable","reason":"El resultado es un análisis, sin código"}]}`, "not_applicable"},
		{`{"checks":[{"name":"editorial","result":"passed","evidence":"Revisión del borrador"},{"name":"fuente","result":"failed","evidence":"La cita contradice la fuente"}]}`, "failed"},
	}
	for _, tc := range cases {
		r, err := Decode(strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}
		if got := Assess(r); got.Conclusion != tc.want {
			t.Fatalf("assessment: %#v want %s", got, tc.want)
		}
	}
}
func TestNotApplicableIsCountedSeparately(t *testing.T) {
	r, err := Decode(strings.NewReader(`{"checks":[{"name":"comportamiento","result":"passed","evidence":"Caso reproducible"},{"name":"coverage","result":"not_applicable","reason":"No es código"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	a := Assess(r)
	if a.Passed != 1 || a.NotApplicable != 1 || a.Conclusion != "supported" {
		t.Fatalf("counts: %#v", a)
	}
}
func TestRejectAmbiguousOrEmptyDeclarationsWithoutLeakingContent(t *testing.T) {
	for _, body := range []string{
		`{"schema_version":"result-contract/v1","status":"validated"}`,
		`{"checks":[]}`,
		`{"Checks":[{"name":"x","result":"passed","evidence":"observado"}]}`,
		`{"checks":[{"name":"x","result":"passed"}]}`,
		`{"checks":[{"name":"x","result":"not_applicable"}]}`,
		`{"checks":[{"name":"x","result":"passed","evidence":"observado","score":5}]}`,
		`{"checks":[{"name":"x","result":"passed","result":"failed","evidence":"observado"}]}`,
		`{"checks":[{"name":"x","result":"passed","evidence":"observado"},{"name":"X","result":"failed","evidence":"observado"}]}`,
		`{"checks":[{"name":"x","result":"passed","evidence":"observado"}],"findings":[{"severity":"extreme","summary":"PRIVATE-CANARY"}]}`,
		`{"checks":[{"name":"x","result":"passed","evidence":"observado"}]} {}`,
	} {
		if _, err := Decode(strings.NewReader(body)); err == nil || strings.Contains(err.Error(), "PRIVATE-CANARY") {
			t.Fatalf("accepted ambiguous input or leaked body: %v", err)
		}
	}
}
func TestFingerprintBindsObservationsButIgnoresJSONFormattingAndOrder(t *testing.T) {
	a, err := Decode(strings.NewReader(`{"checks":[{"name":"B","result":"passed","evidence":"observado B"},{"name":"A","result":"passed","evidence":"observado A"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Decode(strings.NewReader(`{ "checks": [ {"evidence":"observado A","result":"passed","name":"A"},{"result":"passed","name":"B","evidence":"observado B"} ] }`))
	if err != nil {
		t.Fatal(err)
	}
	if Fingerprint(a) != Fingerprint(b) {
		t.Fatal("format or order affected identity")
	}
	b.Checks[0].Evidence = "Otra observación"
	if Fingerprint(a) == Fingerprint(b) {
		t.Fatal("changed observation retained identity")
	}
}
