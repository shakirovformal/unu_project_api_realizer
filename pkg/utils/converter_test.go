package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverterUnfullfilledKeys(t *testing.T) {
	testData := []string{"2", "0", "15", "7"}
	expRes := []int{0, 2, 7, 15}
	gotRes, err := ConverterUnfullfilledKeys(testData)
	fmt.Println(gotRes)
	assert.Equal(t, expRes, gotRes)
	require.NoError(t, err)

}
