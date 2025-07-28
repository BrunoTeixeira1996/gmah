package email

import (
	"os"
	"testing"
)

// Helper function to load HTML files for testing
func loadTestHTMLFile(t *testing.T, filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filename, err)
	}
	return string(content)
}

// Test the processEmailBody function
func TestProcessEmailBody(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		bodyFile    string
		wantLink    string
		wantSnippet string
	}{
		/*		{
					name:        "idealista",
					from:        "idealista",
					bodyFile:    "/home/brun0/Desktop/personal/gmah/testdata/idealista.html",
					wantLink:    `3D"https://www.idealista.pt/imovel/33667017/?xts=3D582068&xto=`,
					wantSnippet: "Apartamento T3 em praceta Doutor Alberto Tavares de Castro 9 Oliveira do Bairro Oliveira do Bairro 160000 E282AC Apartamento T3 venda no Centro da CidadeDescubra este excelente apartamento T3 que co Ver 9 fotos 160000 E282AC 160000 E282AC€",
				},
				{
					name:        "SUPERCASA",
					from:        "SUPERCASA",
					bodyFile:    "/home/brun0/Desktop/personal/gmah/testdata/SUPERCASA.html",
					wantLink:    "https://supercasa.pt/venda-apartamento-t3-aveiro/i1736538?utm_source=scalert&utm_medium=immediatealert-newrealestate&utm_campaign=20240921&mid=583735611&ansid=674057883&euid=mb1EXd64Jg7G1fa2ijnWvA==&ffcf=1",
					wantSnippet: "Apartamento T3 venda em Glria e Vera Cruz",
				},
				{
					name:        "Imovirtual",
					from:        "Imovirtual",
					bodyFile:    "/home/brun0/Desktop/personal/gmah/testdata/imovirtual.html",
					wantLink:    "https://www.imovirtual.com/pt/anuncio/moradia-t3-para-venda-em-anadia-ID1fxx0?utm_medium=email&utm_source=siren&utm_campaign=saved-search-immediate",
					wantSnippet: "Moradia T3 para venda em Anadia",
				},*/
		{
			name:        "CasaYes",
			from:        "CasaYes",
			bodyFile:    "/home/brun0/Desktop/personal/gmah/testdata/casayes.html",
			wantLink:    "https://1818X.trk.elasticemail.com/tracking/click?d=fONHM7NUd7C3oiKWHQrWe2020qp_xSD1KaqL1Eq9CPW9uvjqqkosmRpzWW5Drx3_XGNnBLoGCohIU0mhg794xC4sPGN2EmtfMRQrXcfYOsWnlrbUHLHFo6DV4UzhtEYsWYX8ZE8AmBWlGO3Jq8awf0iVq87gq8OKX7udCRFsd4pF0",
			wantSnippet: "Apartamento T3 Ovar So Joo Arada e So Vicente de Pereira Jus Ovar",
		},
		{
			name:        "CasaYes",
			from:        "CasaYes",
			bodyFile:    "/home/brun0/Desktop/personal/gmah/testdata/casayes2.html",
			wantLink:    "https://1818X.trk.elasticemail.com/tracking/click?d=fONHM7NUd7C3oiKWHQrWe2020qp_xSD1KaqL1Eq9CPX_UpDV5qx3CZN-pebYlv0tUucJkdANC2LAJMPjGuYGOmwN7ptmVsKiitrEG556wiC4cOz7kmoDB1JFJtUnlCq_r-ArFCL0cHCISsALI7N066eFGq9NvJGUWXdT5PJkP87u0",
			wantSnippet: "Moradia T3 Esgueira Aveiro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load the HTML content from the file
			body := loadTestHTMLFile(t, tt.bodyFile)

			email, err := ProcessEmailBody(tt.from, body)
			if err != nil {
				t.Fatalf("processEmailBody() returned an error: %v", err)
			}

			if email.Link != tt.wantLink {
				t.Errorf("expected Link to be %s, got %s", tt.wantLink, email.Link)
			}
			if email.Snippet != tt.wantSnippet {
				t.Errorf("expected Snippet to be %s, got %s", tt.wantSnippet, email.Snippet)
			}
		})
	}
}
