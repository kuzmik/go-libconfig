package libconfig

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkParseSimpleConfig benchmarks parsing a simple configuration.
func BenchmarkParseSimpleConfig(b *testing.B) {
	config := `
		name = "MyApp";
		port = 8080;
		debug = true;
		version = 1.5;
	`

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseComplexConfig benchmarks parsing a complex nested configuration.
func BenchmarkParseComplexConfig(b *testing.B) {
	config := `
		application = {
			name = "Complex App";
			version = "2.1.0";
			debug = false;

			server = {
				host = "localhost";
				port = 8080;
				ssl = {
					enabled = true;
					cert_file = "/etc/ssl/cert.pem";
					key_file = "/etc/ssl/key.pem";
				};
			};

			database = {
				driver = "postgresql";
				connection = {
					host = "db.example.com";
					port = 5432;
					database = "myapp";
					username = "user";
					password = "secret";
				};
				pool = {
					min_connections = 5;
					max_connections = 100;
					idle_timeout = 300.0;
				};
			};

			features = {
				analytics = true;
				caching = false;
				notifications = [ "email", "sms", "push" ];
			};
		};

		logging = {
			level = "INFO";
			outputs = [ "console", "file" ];
			file_config = {
				path = "/var/log/app.log";
				max_size = 104857600;
				rotation = true;
			};
		};
	`

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseLargeArray benchmarks parsing configurations with large arrays.
func BenchmarkParseLargeArray(b *testing.B) {
	// Generate a large array
	var items []string
	for i := 0; i < 1000; i++ {
		items = append(items, fmt.Sprintf(`"item_%d"`, i))
	}

	config := fmt.Sprintf(`large_array = [ %s ];`, strings.Join(items, ", "))

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseManySettings benchmarks parsing many flat settings.
func BenchmarkParseManySettings(b *testing.B) {
	var settings []string
	for i := 0; i < 500; i++ {
		settings = append(settings, fmt.Sprintf(`setting_%d = %d;`, i, i))
	}

	config := strings.Join(settings, "\n")

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseNumbers benchmarks parsing different number formats.
func BenchmarkParseNumbers(b *testing.B) {
	config := `
		decimal = 42;
		hex = 0xFF;
		binary = 0b1010;
		octal = 0o755;
		float = 3.14159;
		scientific = 1.23e-4;
		long_int = 9223372036854775807L;
	`

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseStrings benchmarks parsing string values with various features.
func BenchmarkParseStrings(b *testing.B) {
	config := `
		simple = "Hello World";
		with_escapes = "Line 1\nLine 2\tTabbed";
		with_quotes = "He said, \"Hello there!\"";
		concatenated = "This is a very long string that "
		              "spans multiple lines and demonstrates "
		              "automatic string concatenation.";
		unicode = "Unicode: \x41\x42\x43";
	`

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLookupShallow benchmarks shallow lookup operations.
func BenchmarkLookupShallow(b *testing.B) {
	config, err := ParseString(`
		name = "test";
		port = 8080;
		debug = true;
		version = 1.5;
	`)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		_, err := config.LookupString("name")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLookupDeep benchmarks deep nested lookup operations.
func BenchmarkLookupDeep(b *testing.B) {
	config, err := ParseString(`
		app = {
			database = {
				connection = {
					settings = {
						timeout = 30;
					};
				};
			};
		};
	`)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		_, err := config.LookupInt("app.database.connection.settings.timeout")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLookupTypes benchmarks different type lookup operations.
func BenchmarkLookupTypes(b *testing.B) {
	config, err := ParseString(`
		str_val = "hello";
		int_val = 42;
		float_val = 3.14;
		bool_val = true;
		int64_val = 9223372036854775807L;
	`)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		_, _ = config.LookupString("str_val")
		_, _ = config.LookupInt("int_val")
		_, _ = config.LookupFloat("float_val")
		_, _ = config.LookupBool("bool_val")
		_, _ = config.LookupInt64("int64_val")
	}
}

// BenchmarkArrayAccess benchmarks accessing array elements.
func BenchmarkArrayAccess(b *testing.B) {
	var items []string
	for i := 0; i < 100; i++ {
		items = append(items, fmt.Sprintf(`"item_%d"`, i))
	}

	configStr := fmt.Sprintf(`items = [ %s ];`, strings.Join(items, ", "))

	config, err := ParseString(configStr)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		arrayVal, err := config.Lookup("items")
		if err != nil {
			b.Fatal(err)
		}
		// Access all elements
		for _, item := range arrayVal.ArrayVal {
			_ = item.StrVal
		}
	}
}

// BenchmarkGroupAccess benchmarks accessing group members.
func BenchmarkGroupAccess(b *testing.B) {
	config, err := ParseString(`
		database = {
			host = "localhost";
			port = 5432;
			username = "admin";
			password = "secret";
			ssl_enabled = true;
			timeout = 30.0;
		};
	`)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		dbVal, err := config.Lookup("database")
		if err != nil {
			b.Fatal(err)
		}
		// Access all group members
		for _, value := range dbVal.GroupVal {
			_ = value.Type
		}
	}
}

// BenchmarkCommentParsing benchmarks parsing configurations with many comments.
func BenchmarkCommentParsing(b *testing.B) {
	config := `
		// This is a C++ style comment
		name = "MyApp"; // inline comment

		/* This is a C style comment */
		port = 8080; /* another inline comment */

		# This is a script style comment
		debug = true; # yet another inline comment

		/*
		 * Multi-line comment
		 * with multiple lines
		 * of text
		 */
		version = "1.0";

		// More comments
		host = "localhost"; // with values

		# Final comment
	`

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMixedDataTypes benchmarks parsing configurations with mixed data types.
func BenchmarkMixedDataTypes(b *testing.B) {
	config := `
		strings = [ "red", "green", "blue" ];
		integers = [ 1, 2, 3, 4, 5 ];
		floats = [ 1.1, 2.2, 3.3 ];
		booleans = [ true, false, true ];

		mixed_list = (
			"string",
			42,
			true,
			3.14,
			{ nested = "object"; count = 5; },
			[ "nested", "array" ]
		);

		complex_group = {
			basic_types = {
				name = "test";
				count = 100;
				ratio = 0.75;
				enabled = false;
			};

			collections = {
				tags = [ "important", "urgent", "critical" ];
				settings = (
					{ key = "timeout"; value = 30; },
					{ key = "retries"; value = 3; }
				);
			};
		};
	`

	b.ResetTimer()

	for b.Loop() {
		_, err := ParseString(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValueConstruction benchmarks value constructor functions.
func BenchmarkValueConstruction(b *testing.B) {
	b.ResetTimer()

	for b.Loop() {
		_ = NewIntValue(42)
		_ = NewInt64Value(9223372036854775807)
		_ = NewFloatValue(3.14159)
		_ = NewBoolValue(true)
		_ = NewStringValue("hello world")
		_ = NewArrayValue([]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)})
		_ = NewGroupValue(map[string]Value{"key": NewStringValue("value")})
		_ = NewListValue([]Value{NewStringValue("mixed"), NewIntValue(42)})
	}
}
