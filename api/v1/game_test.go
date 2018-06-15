package v1

import "testing"

func TestGame_getStatus(t *testing.T) {
	type fields struct {
		Board string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "X Horizontal 1 Win",
			fields: fields{
				Board: "XXXOO----",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Horizontal 1 Win",
			fields: fields{
				Board: "OOOXX----",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Horizontal 2 Win",
			fields: fields{
				Board: "---XXXOO-",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Horizontal 2 Win",
			fields: fields{
				Board: "---OOOXX-",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Horizontal 3 Win",
			fields: fields{
				Board: "----OOXXX",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Horizontal 3 Win",
			fields: fields{
				Board: "----XXOOO",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Vertical 1 Win",
			fields: fields{
				Board: "X--X--XOO",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Vertical 1 Win",
			fields: fields{
				Board: "O--O--OXX",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Vertical 2 Win",
			fields: fields{
				Board: "-X--X-OXO",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Vertical 2 Win",
			fields: fields{
				Board: "-O--O-XOX",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Vertical 3 Win",
			fields: fields{
				Board: "--X--XOOX",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Vertical 3 Win",
			fields: fields{
				Board: "--O--OXXO",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Diagonal 1 Win",
			fields: fields{
				Board: "X---X-OOX",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Diagonal 1 Win",
			fields: fields{
				Board: "O---O-XXO",
			},
			want: gameStatusOWon,
		},
		{
			name: "X Diagonal 2 Win",
			fields: fields{
				Board: "--X-XOXO-",
			},
			want: gameStatusXWon,
		},
		{
			name: "O Diagonal 2 Win",
			fields: fields{
				Board: "--O-OXOX-",
			},
			want: gameStatusOWon,
		},
		{
			name: "Draw",
			fields: fields{
				Board: "OXXXOOOOX",
			},
			want: gameStatusDraw,
		},
		{
			name: "Running",
			fields: fields{
				Board: "O--------",
			},
			want: gameStatusRunning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				Board: tt.fields.Board,
			}
			if got := g.getStatus(); got != tt.want {
				t.Errorf("Game.getStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
