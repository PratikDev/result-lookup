package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pratikdev/result-lookup/internal/models"
	"github.com/pratikdev/result-lookup/internal/testutils"
)

func TestGetResult(t *testing.T) {
	// flush miniredis
	testMR.FlushAll()

	// truncate table
	testutils.TruncateResultsTable(testFbDbPool)

	// wrap GetResult func to
	// prevent repeated args
	gr := func (w *httptest.ResponseRecorder, r *http.Request) {
		GetResult(w, r, testFbDbPool, testRDB, testLogger, testValidator)
	}

	t.Run("test invalid request (no search param)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/result", nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected code: %d, got: %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("test invalid request (invalid roll)", func(t *testing.T) {
		testParam := "roll=lkoower"
		r := httptest.NewRequest(http.MethodGet, "/result?reg=1234&exam_year=2010&" + testParam, nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected code: %d, got: %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("test invalid request (invalid reg)", func(t *testing.T) {
		testParam := "reg=lkoower"
		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&exam_year=2010&" + testParam, nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected code: %d, got: %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("test invalid request (invalid exam year - gibberish)", func(t *testing.T) {
		testParam := "exam_year=lkoower"
		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&reg=1234&" + testParam, nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected code: %d, got: %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("test invalid request (invalid exam year - out of range)", func(t *testing.T) {
		testParam := "exam_year=1990"
		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&reg=1234&" + testParam, nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected code: %d, got: %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("test missing publish gate", func(t *testing.T) {
		// flush miniredis
	  testMR.FlushAll()
		
		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&reg=5678&exam_year=2010", nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected code: %d, got: %d", http.StatusServiceUnavailable, w.Code)
		}
	})

	t.Run("test gate is 0", func(t *testing.T) {
		// flush miniredis
	  testMR.FlushAll()

		// set published result to "0"
		testRDB.Set(t.Context(), "result:published", "0", 0)

		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&reg=5678&exam_year=2010", nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected code: %d, got: %d", http.StatusServiceUnavailable, w.Code)
		}
	})

	t.Run("test gate is 1, result missing", func(t *testing.T) {
		// flush miniredis
	  testMR.FlushAll()

		// set published result to "1"
		testRDB.Set(t.Context(), "result:published", "1", 0)

		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&reg=5678&exam_year=2010", nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected code: %d, got: %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("test gate is 1, result exists", func(t *testing.T) {
		// flush miniredis
	  testMR.FlushAll()
		
		// set published result to "1"
		testRDB.Set(t.Context(), "result:published", "1", 0)

		respBody := models.Result{
			ID: 000,
			Roll: 1234,
			Reg: 5678,
			StudentName: "testName",
			InstitutionName: "testInst",
			BoardName: "testBoard",
			ExamYear: 2010,
			GPA: 4.00,
			IsPassed: true,
		}

		jsonBytes, err := json.Marshal(respBody)
		if err != nil {
			t.Error("response body json marshel failed")
			return
		}
	
		resultJsonString := string(jsonBytes)

		// set results
		testRDB.Set(t.Context(), "result:1234:5678:2010", resultJsonString, 0)

		r := httptest.NewRequest(http.MethodGet, "/result?roll=1234&reg=5678&exam_year=2010", nil)
		w := httptest.NewRecorder()

		// ping
		gr(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("expected code: %d, got: %d", http.StatusOK, w.Code)
		}

		var responseBody models.Result
		err = json.NewDecoder(w.Body).Decode(&responseBody)
		if err != nil {
			t.Error("wrong response body shape")
			return
		}

		if responseBody.StudentName != respBody.StudentName {
			t.Errorf("student name mismatch. expected=%s, got=%s", respBody.StudentName, responseBody.StudentName)
		}

		if responseBody.GPA != respBody.GPA {
			t.Errorf("GPA mismatch. expected=%f, got=%f", respBody.GPA, responseBody.GPA)
		}

		if responseBody.IsPassed != respBody.IsPassed {
			t.Errorf("Is-Passed mismatch. expected=%v, got=%v", respBody.IsPassed, responseBody.IsPassed)
		}
	})
}