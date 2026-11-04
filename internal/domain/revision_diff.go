package domain

import "strings"

type DiffOp string

const (
	DiffEqual  DiffOp = "equal"
	DiffAdd    DiffOp = "add"
	DiffRemove DiffOp = "remove"
)

// Vượt ngưỡng này thì không tính LCS nữa, coi như xóa hết bản cũ và thêm hết bản mới.
const maxDiffCells = 4_000_000

type DiffLine struct {
	Op   DiffOp
	Text string
}

type FieldChange struct {
	Field string
	From  string
	To    string
}

type RevisionComparison struct {
	From    RevisionData
	To      RevisionData
	Changes []FieldChange
	Lines   []DiffLine
	Added   int
	Removed int
}

// RevisionDataFromPost biểu diễn bản hiện tại của bài viết như một phiên bản (ID = 0).
func RevisionDataFromPost(post PostData) RevisionData {
	return RevisionData{
		PostID: post.ID, Title: post.Title, Slug: post.Slug, Content: post.Content,
		Status: post.Status, CategoryID: post.CategoryID, UserID: post.AuthorID,
		User: post.Author, CreatedAt: post.UpdatedAt,
	}
}

func CompareRevisions(from, to RevisionData) RevisionComparison {
	result := RevisionComparison{From: from, To: to}
	if from.Title != to.Title {
		result.Changes = append(result.Changes, FieldChange{Field: "title", From: from.Title, To: to.Title})
	}
	if from.Slug != to.Slug {
		result.Changes = append(result.Changes, FieldChange{Field: "slug", From: from.Slug, To: to.Slug})
	}
	if from.Status != to.Status {
		result.Changes = append(result.Changes, FieldChange{Field: "status", From: string(from.Status), To: string(to.Status)})
	}
	if from.CategoryID != to.CategoryID {
		result.Changes = append(result.Changes, FieldChange{
			Field: "category_id", From: uintToString(from.CategoryID), To: uintToString(to.CategoryID),
		})
	}
	result.Lines = DiffLines(from.Content, to.Content)
	for _, line := range result.Lines {
		switch line.Op {
		case DiffAdd:
			result.Added++
		case DiffRemove:
			result.Removed++
		}
	}
	return result
}

// DiffLines so sánh hai đoạn văn theo từng dòng (thuật toán LCS).
func DiffLines(oldText, newText string) []DiffLine {
	a, b := splitDiffLines(oldText), splitDiffLines(newText)

	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix &&
		a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}

	lines := make([]DiffLine, 0, len(a)+len(b))
	for _, text := range a[:prefix] {
		lines = append(lines, DiffLine{Op: DiffEqual, Text: text})
	}
	lines = append(lines, diffMiddle(a[prefix:len(a)-suffix], b[prefix:len(b)-suffix])...)
	for _, text := range a[len(a)-suffix:] {
		lines = append(lines, DiffLine{Op: DiffEqual, Text: text})
	}
	return lines
}

func diffMiddle(a, b []string) []DiffLine {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	lines := make([]DiffLine, 0, len(a)+len(b))
	if len(a) == 0 || len(b) == 0 || len(a)*len(b) > maxDiffCells {
		for _, text := range a {
			lines = append(lines, DiffLine{Op: DiffRemove, Text: text})
		}
		for _, text := range b {
			lines = append(lines, DiffLine{Op: DiffAdd, Text: text})
		}
		return lines
	}

	n, m := len(a), len(b)
	width := m + 1
	// table[i*width+j] = độ dài chuỗi con chung dài nhất của a[i:] và b[j:].
	table := make([]int32, (n+1)*width)
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				table[i*width+j] = table[(i+1)*width+j+1] + 1
			case table[(i+1)*width+j] >= table[i*width+j+1]:
				table[i*width+j] = table[(i+1)*width+j]
			default:
				table[i*width+j] = table[i*width+j+1]
			}
		}
	}

	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			lines = append(lines, DiffLine{Op: DiffEqual, Text: a[i]})
			i++
			j++
		case table[(i+1)*width+j] >= table[i*width+j+1]:
			lines = append(lines, DiffLine{Op: DiffRemove, Text: a[i]})
			i++
		default:
			lines = append(lines, DiffLine{Op: DiffAdd, Text: b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		lines = append(lines, DiffLine{Op: DiffRemove, Text: a[i]})
	}
	for ; j < m; j++ {
		lines = append(lines, DiffLine{Op: DiffAdd, Text: b[j]})
	}
	return lines
}

func splitDiffLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func uintToString(value uint) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 20)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
