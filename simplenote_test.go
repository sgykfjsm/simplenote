package simplenote

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	debug           = true
	email, password string
	contents        = []string{
		"this is 1st test content",
		"this is 2nd test content",
		"this is 3rd test content",
	}
	tags = [][]string{
		{"tag1"},
		{"tag1", "tag2"},
		{"tag1", "tag2", "tag3"},
	}

	updateContent = "this is update content"
)

func init() {
	email = os.Getenv("simplenote_email")
	password = os.Getenv("simplenote_password")
	if email == "" {
		fmt.Print("Email for testing is missing.")
		fmt.Print("Please inject your simplenote email into environment variable 'simplenote_email'")
		os.Exit(1)
	}

	if password == "" {
		fmt.Print("Password for testing is missing.")
		fmt.Print("Please inject your simplenote email into environment variable 'simplenote_password'")
		os.Exit(1)
	}
}

func TestSimplenoteClient(t *testing.T) {
	var s *Client
	var errResponse *ErrorResponse
	t.Run("Create Client", func(t *testing.T) {
		s, errResponse = New(email, password, debug)
		if errResponse != nil {
			t.Logf("Failed to create simplenote client: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}
	})

	t.Run("Create Note", func(t *testing.T) {
		for i := range contents {
			_, errResponse = s.Add(contents[i], tags[i])
			if errResponse != nil {
				t.Logf("Failed to add new note: %s", errResponse.Err.Error())
				t.Logf("Status Code: %d", errResponse.Status())
				t.Fail()
			}
			//time.Sleep(1 * time.Second)
		}
	})

	var index Data
	var data Data
	t.Run("Get Index", func(t *testing.T) {
		since := fmt.Sprintf("%f", float64(time.Now().Add(-1*time.Minute).UnixNano()/int64(time.Second)))
		expectedCount := 3
		data, errResponse = s.Index(expectedCount, since, "")
		if errResponse != nil {
			t.Logf("Failed to get note index: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}
		if data.Count != expectedCount {
			t.Logf("Expcted data.Count should be %d but actual is %d", expectedCount, data.Count)
			t.Fail()
		}

		_, errResponse = s.Index(1, since, data.Data[0].Mark)
		if errResponse != nil {
			t.Logf("Failed to get note index by mark: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}

		index = data
	})

	t.Run("Get Note", func(t *testing.T) {
		data, errResponse = s.Get(data.Data[0].Key)
		if errResponse != nil {
			t.Logf("Failed to get note: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}
		if data.Content != contents[0] {
			t.Logf("Expected content is %q but actual is %q", contents[0], data.Content)
			t.Fail()
		}
	})

	var updateData Data
	t.Run("Update Note", func(t *testing.T) {
		updateData, errResponse = s.Update(data.Key, updateContent, false)
		if errResponse != nil {
			t.Logf("Failed to update note: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}
		if updateData.Version != data.Version+1 {
			t.Logf("Expected version is %d but actual version is %d", data.Version, updateData.Version)
			t.Fail()
		}
	})

	var deletedData Data
	t.Run("Mark Note As Deleted", func(t *testing.T) {
		deletedData, errResponse = s.Update(data.Key, "", true)
		if errResponse != nil {
			t.Logf("Failed to update note: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}
		if deletedData.Deleted != 1 {
			t.Log("The content should be deleted but not.")
			t.Fail()
		}
	})

	t.Run("Delete Note", func(t *testing.T) {
		errResponse = s.Delete(deletedData.Key)
		if errResponse != nil {
			t.Logf("Failed to update note: %s", errResponse.Err.Error())
			t.Logf("Status Code: %d", errResponse.Status())
			t.Fail()
		}
		_, errResponse = s.Get(deletedData.Key)
		if errResponse.Status() != http.StatusNotFound {
			t.Logf("Expected status is %q but actiaul is %q",
				http.StatusText(http.StatusNotFound), http.StatusText(errResponse.Status()))
			t.Fail()
		}
	})

	// cleanup
	log.Print(">>> Cleanup")
	for _, data := range index.Data {
		s.Update(data.Key, "", true)
		s.Delete(data.Key)
	}
}
