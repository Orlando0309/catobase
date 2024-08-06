// catobase_test.go
package catobase

import (
	"bufio"
	"os"
	"testing"
)

// Helper function to set up a test file with initial content.
func setupTestFile(t *testing.T, fileName string, content []string) {
	t.Helper()
	file, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer file.Close()
	for _, line := range content {
		if _, err := file.WriteString(line + "\n"); err != nil {
			t.Fatalf("failed to write to test file: %v", err)
		}
	}
}

// Helper function to clean up test files.
func cleanupTestFile(t *testing.T, fileName string) {
	t.Helper()
	if err := os.Remove(fileName); err != nil {
		t.Fatalf("failed to clean up test file: %v", err)
	}
}

func TestCreateCategory(t *testing.T) {
	fileName := "test_create_category.txt"
	defer cleanupTestFile(t, fileName)

	// Test creating a new category file
	success, err := CreateCategory([]string{"Books", "Movies"}, fileName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !success {
		t.Errorf("expected success, got failure")
	}

	// Test creating a file that already exists
	_, err = CreateCategory([]string{"Books", "Movies"}, fileName)
	if err == nil || err.Error() != "file already exists" {
		t.Errorf("expected error 'file already exists', got %v", err)
	}
}

func TestDeleteCategory(t *testing.T) {
	fileName := "test_delete_category.txt"
	setupTestFile(t, fileName, []string{"Books", "Movies", "Music"})
	defer cleanupTestFile(t, fileName)

	// Test deleting an existing category
	success, err := DeleteCategory("Movies", fileName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !success {
		t.Errorf("expected success, got failure")
	}

	// Check if the category is deleted
	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		if scanner.Text() == "Movies" {
			found = true
			break
		}
	}
	if found {
		t.Errorf("expected 'Movies' to be deleted")
	}

	// Test deleting a non-existent category
	success, err = DeleteCategory("NonExistent", fileName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !success {
		t.Errorf("expected success, got failure")
	}
}

func TestFormat(t *testing.T) {
	// Test with default separator
	result := format("/path/to/file", []string{"Books", "Movies"}, "")
	expectedPrefix := "/path/to/file|Books,Movies|"

	if !matchesFormat(result, expectedPrefix) {
		t.Errorf("expected format prefix: %s, got: %s", expectedPrefix, result)
	}

	// Test with custom separator
	result = format("/path/to/file", []string{"Books", "Movies"}, "#")
	expectedPrefix = "/path/to/file#Books,Movies#"

	if !matchesFormat(result, expectedPrefix) {
		t.Errorf("expected format prefix: %s, got: %s", expectedPrefix, result)
	}
}

// Helper function to check if the formatted result matches the expected format.
func matchesFormat(result, expectedPrefix string) bool {
	if len(result) <= len(expectedPrefix) {
		return false
	}
	return result[:len(expectedPrefix)] == expectedPrefix
}

func TestRegisterFile(t *testing.T) {
    testFile := "test_register_file.txt"
    dbFile := ".catodb"
    setupTestFile(t, testFile, []string{"Books", "Movies"})
    defer cleanupTestFile(t, testFile)

    // Create the database file
    db, err := os.Create(dbFile)
    if err != nil {
        t.Fatalf("failed to create .catodb file: %v", err)
    }
    db.Close()
    defer cleanupTestFile(t, dbFile)

    // Test registering a file with existing categories
    success, err := registerFile(testFile, []string{"Books", "Movies"}, true)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if !success {
        t.Errorf("expected success, got failure")
    }

    // Check if the record is written to the database file
    db, err = os.Open(dbFile)
    if err != nil {
        t.Fatalf("failed to open .catodb file: %v", err)
    }
    defer db.Close()

    found := false
    scanner := bufio.NewScanner(db)
    for scanner.Scan() {
        line := scanner.Text()
        if line != "" {
            found = true
            break
        }
    }
    if !found {
        t.Errorf("expected a record in .catodb, found none")
    }

    // Test with non-existing categories
    success, err = registerFile(testFile, []string{"NonExistent"}, false)
    if err == nil || err.Error() != "some categories do not exist" {
        t.Errorf("expected error 'some categories do not exist', got %v", err)
    }
    if success {
        t.Errorf("expected failure, got success")
    }
}


func TestReadFile(t *testing.T) {
	fileName := "test_read_file.txt"
	content := []string{"Books", "Movies", "Music"}
	setupTestFile(t, fileName, content)
	defer cleanupTestFile(t, fileName)

	lines, err := readFile(fileName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(lines) != len(content) {
		t.Errorf("expected %d lines, got %d", len(content), len(lines))
	}

	for i, line := range lines {
		if line != content[i] {
			t.Errorf("expected line %d to be %q, got %q", i, content[i], line)
		}
	}
}

func TestRegisterFiles(t *testing.T) {
	folder := "test_folder"
	os.Mkdir(folder, 0755)
	defer os.RemoveAll(folder)

	file1 := folder + "/file1.txt"
	file2 := folder + "/file2.txt"
	setupTestFile(t, file1, []string{"Books", "Movies"})
	setupTestFile(t, file2, []string{"Music", "Games"})
	
	dbFile := ".catodb"
	db, err := os.Create(dbFile)
	if err != nil {
		t.Fatalf("failed to create .catodb file: %v", err)
	}
	db.Close()
	defer cleanupTestFile(t, dbFile)

	registeredFiles, err := registerFiles(folder, "file.*\\.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(registeredFiles) != 2 {
		t.Errorf("expected 2 registered files, got %d", len(registeredFiles))
	}
}

func TestGet(t *testing.T) {
	dbFile := ".catodb"
	setupTestFile(t, dbFile, []string{
		"/path/to/file1|Books,Movies|2023-07-01T00:00:00Z",
		"/path/to/file2|Music,Games|2023-07-01T00:00:00Z",
	})
	defer cleanupTestFile(t, dbFile)

	matches, err := get("file1", []string{"Books"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}

	if matches[0] != "/path/to/file1" {
		t.Errorf("expected match to be /path/to/file1, got %s", matches[0])
	}
}
