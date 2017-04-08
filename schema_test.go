package sqlrow

import (
	"testing"
)

func TestSchemaDefaults(t *testing.T) {
	dialect1 := DialectFor("postgres")
	dialect2 := DialectFor("ql")
	convention1 := ConventionSame
	convention2 := ConventionSnake

	tests := []struct {
		defaultConvention  Convention
		defaultDialect     Dialect
		schemaConvention   Convention
		schemaDialect      Dialect
		expectedConvention Convention
		expectedDialect    Dialect
	}{
		{
			defaultConvention:  convention1,
			defaultDialect:     dialect1,
			expectedConvention: convention1,
			expectedDialect:    dialect1,
		},
		{
			defaultConvention:  convention2,
			defaultDialect:     dialect1,
			schemaConvention:   convention1,
			schemaDialect:      dialect2,
			expectedConvention: convention1,
			expectedDialect:    dialect2,
		},
	}

	resetDefaults := func() {
		Default = &Schema{}
	}
	defer resetDefaults()

	for _, tt := range tests {
		resetDefaults()
		Default.Convention = tt.defaultConvention
		Default.Dialect = tt.defaultDialect
		schema := &Schema{
			Convention: tt.schemaConvention,
			Dialect:    tt.schemaDialect,
		}
		if schema.convention().Convert("XyzAbc") != tt.expectedConvention.Convert("XyzAbc") {
			t.Errorf("unexpected convention: %v, %v", schema.convention(), tt.expectedConvention)
		}
		if schema.dialect().Name() != tt.expectedDialect.Name() {
			t.Error("unexpected dialect")
		}
	}
}
