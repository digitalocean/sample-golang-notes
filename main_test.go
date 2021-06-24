package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/digitalocean-apps/sample-with-database/pkg/storer"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	uuidRegex = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
)

func TestNotesHandler(t *testing.T) {
	tcs := []struct {
		name           string
		requestMethod  string
		requestBody    string
		mock           func(m sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   "POST a note",
		},
		{
			name:           "POST empty",
			requestMethod:  "POST",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "invalid empty note",
		},
		{
			name:          "POST with body",
			requestMethod: "POST",
			requestBody:   "foobar",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "notes" ("created_at","updated_at","deleted_at","uuid","body") VALUES ($1,$2,$3,$4,$5) RETURNING "notes"."id"`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, sqlmock.AnyArg(), "foobar").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   uuidRegex.String(),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.requestMethod, "/", strings.NewReader(tc.requestBody))
			require.NoError(t, err)

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			gdb, err := gorm.Open("postgres", db)
			require.NoError(t, err)
			gdb.LogMode(true)

			if tc.mock != nil {
				tc.mock(mock)
			}

			storerClient := &storer.PG{
				DB: gdb,
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(notesHandler(storerClient))
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			r := regexp.MustCompile(tc.expectedBody)
			body := rr.Body.String()
			assert.True(t, r.MatchString(body), fmt.Sprintf("%s does not match %s", body, r.String()))

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

func TestNoteHandler(t *testing.T) {
	tcs := []struct {
		name           string
		noteID         string
		mock           func(m sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET empty arg",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "note_id required",
		},
		{
			name:           "GET not found",
			expectedStatus: http.StatusNotFound,
			noteID:         "note-uuid",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notes"  WHERE "notes"."deleted_at" IS NULL AND ((uuid = $1)) LIMIT 1`)).
					WithArgs("note-uuid").
					WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name:   "GET found",
			noteID: "note-uuid",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notes"  WHERE "notes"."deleted_at" IS NULL AND ((uuid = $1)) LIMIT 1`)).
					WithArgs("note-uuid").
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "body"}).AddRow("note-uuid", "foobar"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "foobar",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/%s", "note-uuid"), nil)
			require.NoError(t, err)

			req = mux.SetURLVars(req, map[string]string{
				"note_id": tc.noteID,
			})

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			gdb, err := gorm.Open("postgres", db)
			require.NoError(t, err)
			gdb.LogMode(true)

			if tc.mock != nil {
				tc.mock(mock)
			}

			storerClient := &storer.PG{
				DB: gdb,
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(noteHandler(storerClient))
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, tc.expectedBody, rr.Body.String())

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
