package catobase

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CreateCategory creates a file with the given fileName and writes the categories into it.
// It returns true if the operation is successful, and an error if something goes wrong.
func CreateCategory(categories []string, fileName string) (bool, error) {
    f, err := os.OpenFile(fileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
    if err != nil {
        if errors.Is(err, os.ErrExist) {
            return false, errors.New("file already exists")
        }
        return false, err
    }
    defer f.Close()

    for _, category := range categories {
        if _, err := f.WriteString(category + "\n"); err != nil {
            return false, err
        }
    }
    return true, nil
}

func checkFileExists(filename string) (*os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return file, errors.New("file does not exist")
        }
        return file, err
    }
	return file, nil
}

func DeleteCategory(categoryToRemove string, fileName string) (bool, error) {
    // Open the file for reading
    file, err := checkFileExists(fileName)
    if err != nil {
        return false, err
    }
    defer file.Close()

    // Read the existing categories
    var categories []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        category := scanner.Text()
        if category != categoryToRemove {
            categories = append(categories, category)
        }
    }
    if err := scanner.Err(); err != nil {
        return false, err
    }

    // Write the categories back to the file (overwrite the file)
    f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0644)
    if err != nil {
        return false, err
    }
    defer f.Close()

    for _, category := range categories {
        if _, err := f.WriteString(category + "\n"); err != nil {
            return false, err
        }
    }
    return true, nil
}

func format(path string, categories []string, separator string) string {
    if separator == "" {
        separator = "|"
    }
    cat := ""
    for index, category := range categories {
        cat += category
        if index+1 < len(categories) {
            cat += ","
        }
    }

    t := time.Now()

    return fmt.Sprintf("%s%s%s%s%s", path, separator, cat, separator, t.Format(time.RFC3339))
}

func readFile(fileName string) ([]string, error) {
    file, err := os.Open(fileName)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return lines, nil
}


func registerFile(fileName string, categories []string, copy bool) (bool, error) {
    // Check if the file exists
    file, err := checkFileExists(fileName)
    if err != nil {
        return false, err
    }
    defer file.Close()

    // Format the registration entry
    formatted := format(fileName, categories, "")

    // Open the .catodb file with the correct flags for appending data
    db, err := os.OpenFile(".catodb", os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return false, fmt.Errorf("failed to open .catodb for writing: %w", err)
    }
    defer db.Close()

    // If copy is true, create a copy of the file
    if copy {
        destinationFile := fileName + ".copy"
        dst, err := os.Create(destinationFile)
        if err != nil {
            return false, fmt.Errorf("failed to create copy of file: %w", err)
        }
        defer dst.Close()

        if _, err := io.Copy(dst, file); err != nil {
            return false, fmt.Errorf("failed to copy file: %w", err)
        }
    }

    // Write the formatted entry to .catodb
    _, err = db.WriteString(formatted + "\n")
    if err != nil {
        return false, fmt.Errorf("failed to write to .catodb: %w", err)
    }

    return true, nil
}

// registerFiles scans the folder and registers all files that match the regex.
func registerFiles(folder string, regex string) ([]string, error) {
    var registeredFiles []string
    re, err := regexp.Compile(regex)
    if err != nil {
        return nil, err
    }

    err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && re.MatchString(info.Name()) {
            categories, err := readFile(path)
            if err != nil {
                return err
            }
            success, err := registerFile(path, categories, true)
            if err != nil {
                return err
            }
            if success {
                registeredFiles = append(registeredFiles, path)
            }
        }
        return nil
    })
    if err != nil {
        return nil, err
    }

    return registeredFiles, nil
}

// get searches for registered files based on regex and categories.
func get(regex string, categories []string) ([]string, error) {
    db, err := checkFileExists(".catodb")
    if err != nil {
        return nil, err
    }
    defer db.Close()

    re, err := regexp.Compile(regex)
    if err != nil {
        return nil, err
    }

    var matches []string
    scanner := bufio.NewScanner(db)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, "|")
        if len(parts) < 3 {
            continue
        }
        path := parts[0]
        fileCategories := strings.Split(parts[1], ",")
        if re.MatchString(path) && containsAll(fileCategories, categories) {
            matches = append(matches, path)
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return matches, nil
}

// containsAll checks if all elements of subset are in set.
func containsAll(set, subset []string) bool {
    setMap := make(map[string]struct{}, len(set))
    for _, s := range set {
        setMap[s] = struct{}{}
    }
    for _, s := range subset {
        if _, ok := setMap[s]; !ok {
            return false
        }
    }
    return true
}
