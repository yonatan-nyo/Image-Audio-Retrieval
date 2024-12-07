package types

type Matrix struct {
    data [][]float64
}

func (m *Matrix) Rows() int {
    return len(m.data)
}

func (m *Matrix) Cols() int {
    return len(m.data[0])
}

func (m *Matrix) GetRow(index int) []float64 {
    return m.data[index]
}

func NewMatrix(data [][]float64) *Matrix {
    return &Matrix{
        data: data,
    }
}

func (m *Matrix) Get(index int, jndex int) float64 {
    return m.data[index][jndex]
}

func (m *Matrix) ToSlice() [][]float64 {
    return m.data
}