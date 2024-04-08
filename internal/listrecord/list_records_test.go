package listRecordImplementation

import (
	"path/filepath"
	"testing"
)

const (
	rootDir = "../.."
)

func TestListRecordsWithSamsungData(t *testing.T) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		t.Error(err)
	}
	testDataDir := filepath.Join(rootDir, "test-data")
	samsungTestData := filepath.Join(testDataDir, "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")

	records, err := ListRecords(samsungTestData)
	if err != nil {
		t.Error(err)
	}
	expectedResults := []string{
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:d3aae465-714f-4aa8-8f1b-23e75b09af42",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:a861f483-f5eb-4c56-8246-2938d659cbef",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:20add669-b781-4c66-977c-e50e280a69e9",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:796f1780-1774-4136-8932-5ec0905d1194",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:cf1d5825-7ee2-4541-bbfa-93b265218839",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:21a3e842-8499-4cfd-afa3-99dab6b220bc",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:60331c1b-c2f4-486a-a14f-bd448ba6e1c7",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:7b8ec54b-b6e4-47a9-852c-5520fd17d2b9",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:67dfefea-7a0c-449f-bc0c-92bcfe156381",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:370fd75d-5c2e-45b1-b2a6-b262200e663b",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:d0bb0b80-1b5e-4a72-947f-e80360b55d18",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:27747b18-0ca0-43cd-b238-c7d8d6220579",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:f0c247df-335f-43ed-a6db-4e0c64e69ba5",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:bb637a51-071c-4b22-9fea-7c4b16d32c5f",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:bce354e8-6222-4e3c-a40a-4d0b4aad922d",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:76af6549-ff8e-42eb-8972-342eafa0375f",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:86758469-b8be-4b5b-8c24-4ff13b04e4f1",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:b368e7d4-5454-4266-b739-36bc49f37fdf",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:e51361a4-1de6-4a7e-aedb-74908b071baf",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:e64e17db-b84e-4e84-9c21-13ff3e1b476a",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:a2a21295-8b55-4ce3-95e4-140c34ed65a0",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:d684fe64-c954-4a09-91f2-485f73160346",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:ff4eafdc-3eb9-4a67-84d9-0421ec3f1331",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:85961714-c988-4647-90a7-3d46a20bb884",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:17b9e6f4-24c7-44ba-bea3-7f08799e544c",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:1d9314ae-0a32-4fb2-b687-d5713777f364",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:227df586-e0e9-4cdc-9b4f-90b8be0752cd",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:7c04c3ff-3ec7-46cb-af90-1091226e4bb1",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:d6aa09b6-2ca6-47ed-aa93-28e1e77e96e5",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:867f34cf-960b-409a-9b4a-8f9e812c3196",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:8e62465a-7a48-43da-8e00-1055c063c5f5",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:b02fac9b-12f2-4d40-80fc-e434c4566e28",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:2f1a33f4-6a02-4a2d-bf1e-f9eb4b5d12db",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:13413402-07d6-4e47-a9ca-8c08a8175e6a",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:60b64e91-2809-4f68-b3b1-09e4e3cca01c",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:ea94436c-a829-46af-82dd-c48294c1c318",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:afa93317-20c0-4cd8-9634-46d1f877b775",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:95b0dd7f-cf4b-446f-af81-3031d1cc7a72",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:7a47d16d-2b45-47c4-ab6d-512834c31326",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:d650eff0-610f-456a-bc07-d5486e2081a6",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:8131e2c6-6aed-40e8-8c4a-dab87319413a",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:5d70c59d-680b-4261-902b-922374d698c1",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:5e3381dd-5ac7-461a-9498-0a6a890de7fd",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:2d39a767-af91-4b5c-a09b-cbcfe9b76512",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:cc518da7-12b1-4c31-8a9c-351080821b5c",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:4df61ca6-fdcb-48ae-b5e9-f942681ef53b",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:3a90546d-b788-406d-9b4b-dd2e66ecd449",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:dfab799a-ffb5-451f-b924-c103771688d7",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:e1c1e9bf-5c85-4575-89ac-c0297e31b826",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:6df5f61f-12ea-4407-836e-0655672e9188",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:6330cf8a-a66d-4e37-8658-9609998cd65f",
		"WARC record: version: WARC/1.1, type: request, id: urn:uuid:e80a9a2a-8169-4066-8923-b1ba92b785d0",
		"WARC record: version: WARC/1.1, type: response, id: urn:uuid:a4b63c93-5cea-499e-b670-600ae723d08a",
	}
	recordsAsStrings := make([]string, len(records))

	for index, warcRecord := range records {
		recordsAsStrings[index] = warcRecord.Record.String()
	}

	if len(records) != len(expectedResults) {
		t.Errorf("Expected '%d' records, got '%d'", len(expectedResults), len(records))
	}

	for index, warcRecordAsString := range recordsAsStrings {
		if warcRecordAsString != expectedResults[index] {
			t.Errorf("Expected '%s', got '%s' for index '%d'", expectedResults[index], warcRecordAsString, index)
		}
	}
}
