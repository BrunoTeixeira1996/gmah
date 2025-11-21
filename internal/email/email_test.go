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
		{
			name:        "CasaYes",
			from:        "CasaYes",
			bodyFile:    "testdata/casayes.html",
			wantLink:    "https://1818X.trk.elasticemail.com/tracking/click?d=fONHM7NUd7C3oiKWHQrWe2020qp_xSD1KaqL1Eq9CPWIVpE5FFIyajIWdICR_03s1MvciFtyL9f9F6j4doK9qtmzGYDvb9esNEZHCp0MDObfFnFaeLhd5-X8TxYrhrv92E6GPXLSGQZAaoC8OFii4by5zx6IXf7TpwEOD-G9-zER0",
			wantSnippet: "Moradia T3 Palhaça, Oliveira do Bairro",
		},
		{
			name:        "SUPERCASA",
			from:        "SUPERCASA",
			bodyFile:    "testdata/supercasa.html",
			wantLink:    "https://supercasa.pt/venda-apartamento-t2-aveiro/i2090701?utm_source=scalert&utm_medium=immediatealert-newrealestate&utm_campaign=20251119&mid=1004923422&ansid=1125771351&euid=mb1EXd64Jg7G1fa2ijnWvA==&ffcf=1",
			wantSnippet: "Apartamento T2 à venda em Glória e Vera Cruz",
		},
		{
			name:        "idealista",
			from:        "idealista",
			bodyFile:    "testdata/idealista.html",
			wantLink:    "https://www.idealista.pt/imovel/34643108/?utm_medium=email&utm_campaign=express_newAd_sale_professional&utm_campaigntype=retention&utm_project=adNotification&utm_source=alerts-id&utm_date_send=2025-11-17T134218&utm_link=propertyNewPhoto&utm_recipient_id=x1cRuf6NFjkKjOvji/2z24ei8vg87DfIrqhqOx5Y8RI=&utm_notification_id=f2523200-4d30-46c5-a138-d0125fb8d581",
			wantSnippet: "Moradia geminada em rua Jaime Afreixo, São João da Madeira, São João da Madeira 180000€",
		},
		{
			name:        "Imovirtual",
			from:        "Imovirtual",
			bodyFile:    "testdata/imovirtual.html",
			wantLink:    "https://www.imovirtual.com/pt/anuncio/moradia-para-recuperar-no-centro-de-ilhavo-ID1grL7?utm_medium=email&utm_source=siren&utm_campaign=saved-search-immediate",
			wantSnippet: "Moradia para recuperar no centro de ílhavo",
		},
		/*		{
				name:        "Casa Sapo",
				from:        "Casa Sapo",
				bodyFile:    "testdata/casasapo.html",
				wantLink:    "https://casa.sapo.pt/detalhes.aspx?UID=435e72be-2175-11f0-8add-060000000054&utm_medium=email&utm_campaign=alerts&utm_source=details",
				wantSnippet: "Moradia para recuperar no centro de ílhavo",
			},*/
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
