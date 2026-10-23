package event

import "testing"

func TestSubmitAvailabilityRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		block   TimeBlock
		wantErr bool
	}{
		{
			name: "valid block",
			block: TimeBlock{
				Date:      "2026-10-23",
				StartTime: "09:00",
				EndTime:   "10:30",
			},
		},
		{
			name: "invalid date",
			block: TimeBlock{
				Date:      "23-10-2026",
				StartTime: "09:00",
				EndTime:   "10:30",
			},
			wantErr: true,
		},
		{
			name: "invalid start time",
			block: TimeBlock{
				Date:      "2026-10-23",
				StartTime: "9 AM",
				EndTime:   "10:30",
			},
			wantErr: true,
		},
		{
			name: "invalid end time",
			block: TimeBlock{
				Date:      "2026-10-23",
				StartTime: "09:00",
				EndTime:   "25:00",
			},
			wantErr: true,
		},
		{
			name: "equal times",
			block: TimeBlock{
				Date:      "2026-10-23",
				StartTime: "09:00",
				EndTime:   "09:00",
			},
			wantErr: true,
		},
		{
			name: "end before start",
			block: TimeBlock{
				Date:      "2026-10-23",
				StartTime: "11:00",
				EndTime:   "10:30",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := SubmitAvailabilityRequest{TimeBlocks: []TimeBlock{test.block}}
			err := request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
