package fetch

import "testing"

var articleLookupURLCases = []struct {
	name    string
	baseURL string
	surface string
	query   string
	want    string
	wantErr string
}{
	{
		name:    "builds duda linguistica URL",
		baseURL: "https://example.invalid/dpd",
		surface: "duda-linguistica",
		query:   "cuando-se-escriben-con-tilde-los-adverbios-en-mente",
		want:    "https://example.invalid/duda-linguistica/cuando-se-escriben-con-tilde-los-adverbios-en-mente",
	},
	{
		name:    "builds espanol al dia URL",
		baseURL: "https://example.invalid/dpd",
		surface: "espanol-al-dia",
		query:   "el-adverbio-solo",
		want:    "https://example.invalid/espanol-al-dia/el-adverbio-solo",
	},
	{
		name:    "rejects empty base URL",
		baseURL: " ",
		surface: "espanol-al-dia",
		query:   "solo",
		wantErr: "espanol-al-dia base URL is empty",
	},
}

func TestBuildArticleLookupURL(t *testing.T) {
	for _, tt := range articleLookupURLCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildArticleLookupURL(tt.baseURL, tt.surface, tt.query)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("buildArticleLookupURL() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildArticleLookupURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildArticleLookupURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
