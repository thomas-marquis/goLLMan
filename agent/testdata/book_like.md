---
title: "A Tale of Many Formats"
author: "J. Doe"
date: 2025-08-12
---

# A Tale of Many Formats

This is the introduction paragraph. It should not be considered a heading. It sets the stage for a long text that includes many different markdown constructs and is intended for testing text splitting logic.

## Table of Contents

- Chapter 1
  - Motivation
  - Setup
  - Samples
- Chapter 2
  - Deep Dive
  - Case Studies
- Chapter 3
  - Cross-cutting Concerns
  - Appendices
- Appendix

> Blockquote with some commentary.
> It spans multiple lines.
> It might include a link like https://example.com and some inline `code`.

---

### Chapter 1: Getting Started

Here is a paragraph with a [link](https://example.com) and an image:

![Alt text](image.png "Optional title")

We also have inline code like `fmt.Println("hello")` and bold/italic text.

#### Motivation

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus aliquet, nisl at dictum varius, neque elit facilisis arcu, vitae tincidunt ipsum augue ac nibh. Sed sed malesuada lectus, et efficitur magna.

- Why this book
  - Practical value
    - Real-world examples
    - Clear explanations
  - Theoretical background
- Who should read this
  - Beginners
  - Intermediate users
  - Advanced users

#### Setup

1. Install the necessary tools
   1. Go 1.24
   2. Git
   3. Your favorite editor
2. Clone the repository
   - Using HTTPS
   - Using SSH
3. Run the first example

### Chapter 2: A loooong code bloc

```
import re
from unittest.mock import MagicMock

import pandas as pd
import pytest

from easy_testing.builders import DataFrameBuilder
from easy_testing.dataframes import (
    AssertFrame,
    assert_called_once_with_frame,
    assert_contains_line,
    assert_contains_lines,
    assert_frame_equals,
    assert_frame_partially_equals,
)


class TestAssertCalledOnceWithFrame:
    def test_should_assert_mock_called_once_with_dataframe(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(df)

        # Then
        assert_called_once_with_frame(mock.my_method, df)

    def test_should_assert_mock_called_once_with_dataframe_and_assert_frame_equals_kwargs(self):
        # Given
        mock = MagicMock()
        actual_df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        actual_df.astype({"age": "float64", "job": "string"})

        expected_df = pd.DataFrame(
            [
                ("lolo", 13, "photograph"),
                ("toto", 12, "developer"),
            ],
            columns=["name", "age", "job"],
        )
        expected_df.astype({"age": "int32", "job": "category"})

        # When
        mock.my_method(actual_df)

        # Then
        assert_called_once_with_frame(
            mock.my_method, AssertFrame(expected_df, check_dtype=False, check_row_order=False)
        )

    def test_should_handle_multiple_args(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(1, df, "toto")

        # Then
        assert_called_once_with_frame(mock.my_method, 1, df, "toto")

    def test_should_handle_multiple_args_and_kwargs(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(1, df, "toto", a=1, b=df)

        # Then
        assert_called_once_with_frame(mock.my_method, 1, df, "toto", a=1, b=AssertFrame(df))

    def test_should_raise_assertion_error_when_kwargs_df_values_mismatch(self):
        # Given
        mock = MagicMock()
        actual_df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        expected_df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(b=actual_df, c=2)

        # Then
        with pytest.raises(AssertionError):
            assert_called_once_with_frame(mock.my_method, b=expected_df, c=2)

    def test_should_raise_assertion_error_when_non_df_kwarg_value_mismatch(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(b=df, c=2)

        # Then
        with pytest.raises(AssertionError, match=re.escape("Expected c=3 but got c=2")):
            assert_called_once_with_frame(mock.my_method, b=df, c=3)

    def test_should_raise_assertion_error_when_mock_not_called(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(AssertionError, match=re.escape("Expected to be called once but was called 0 times")):
            assert_called_once_with_frame(mock.my_method, df)

    def test_should_raise_assertion_error_when_not_same_args_number(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(df)

        # Then
        with pytest.raises(AssertionError, match=re.escape("Expected 2 argument(s) but got 1")):
            assert_called_once_with_frame(mock.my_method, df, "toto")

    def test_should_raise_assertion_error_when_mismatch_kwargs_number_an_keys(self):
        # Given
        mock = MagicMock()
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        mock.my_method(1, a=df, b=1)

        # Then
        with pytest.raises(AssertionError, match=re.escape("Expected keyword argument(s) x, y, z but got a, b")):
            assert_called_once_with_frame(mock.my_method, 1, x=df, y=1, z=2)


class TestAssertContainsLine:
    def test_should_return_none_when_given_line_is_present(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        res = assert_contains_line(df, ("toto", 12, "developer"))

        # Then
        assert res is None

    def test_should_return_none_when_given_line_is_present_multiple_times(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
                ("toto", 12, "developer"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        res = assert_contains_line(df, ("toto", 12, "developer"))

        # Then
        assert res is None

    def test_should_raise_when_given_line_is_not_present(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError, match=re.escape("Expected line <name=toto, age=12, job=photograph> not found in dataframe")
        ):
            assert_contains_line(df, ("toto", 12, "photograph"))

    def test_should_raise_when_given_line_is_not_present_at_all(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError, match=re.escape("Expected line <name=tata, age=45, job=seller> not found in dataframe")
        ):
            assert_contains_line(df, ("tata", 45, "seller"))

    def test_should_raise_when_dataframe_is_empty(self):
        # Given
        df = pd.DataFrame([], columns=["name", "age", "job"])

        # When & Then
        with pytest.raises(
            AssertionError, match=re.escape("Expected line <name=tata, age=45, job=seller> not found in dataframe")
        ):
            assert_contains_line(df, ("tata", 45, "seller"))

    def test_should_raise_value_error_when_line_size_mismatch(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(ValueError, match=re.escape("Line size mismatch: expected 3, got 2")):
            assert_contains_line(df, ("tata", 45))


class TestAssertContainsLines:
    def test_should_return_none_when_all_lines_are_present(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        res = assert_contains_lines(df, [("toto", 12, "developer"), ("lolo", 13, "photograph")])

        # Then
        assert res is None

    def test_should_raise_when_one_line_is_not_present(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError, match=re.escape("Expected line <name=tata, age=45, job=seller> not found in dataframe")
        ):
            assert_contains_lines(df, [("toto", 12, "developer"), ("tata", 45, "seller")])

    def test_should_raise_when_multiple_lines_are_not_present(self):
        # Given
        df = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
                ("michou", 23, "soldier"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError,
            match=re.escape(
                "Expected lines <name=tata, age=45, job=seller>, <name=joe, age=34, job=driver> not found in dataframe"
            ),
        ):
            assert_contains_lines(df, [("tata", 45, "seller"), ("joe", 34, "driver")])


class TestAssertPartialFrameEquals:
    def test_should_assert_2_df_are_equals_for_given_columns(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("toto", 12, "cop"),
                ("lolo", 13, "doctor"),
            ],
            columns=["name", "age", "job"],
        )

        # When
        res = assert_frame_partially_equals(df1, df2, ["name", "age"])

        # Then
        assert res is None

    def test_should_assert_2_df_are_not_equals_for_given_columns(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("toto", 12, "cop"),
                ("lolo", 13, "doctor"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_partially_equals(df1, df2, ["age", "job"])

    def test_should_raise_when_given_columns_are_not_present_in_left_df(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "anciennete", "travail"],
        )
        df2 = pd.DataFrame(
            [
                ("toto", 12, "cop"),
                ("lolo", 13, "doctor"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            ValueError,
            match=r"Column\(s\) ('age'|'job'), ('age'|'job') not found in left dataframe",
        ):
            assert_frame_partially_equals(df1, df2, ["age", "job"])

    def test_should_raise_when_given_columns_are_not_present_in_right_df(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("toto", 12, "cop"),
                ("lolo", 13, "doctor"),
            ],
            columns=["name", "anciennete", "travail"],
        )

        # When & Then
        with pytest.raises(
            ValueError,
            match=r"Column\(s\) ('age'|'job'), ('age'|'job') not found in right dataframe",
        ):
            assert_frame_partially_equals(df1, df2, ["name", "age", "job"])

    def test_should_not_raise_when_column_type_ignored(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("toto", 12.0, "cop"),
                ("lolo", 13.0, "doctor"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        assert_frame_partially_equals(df1, df2, ["name", "age"], check_dtype=False)

    def test_should_raise_when_column_type_not_ignored(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("toto", 12.0, "cop"),
                ("lolo", 13.0, "doctor"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_partially_equals(df1, df2, ["name", "age"], check_dtype=True)

    def test_should_not_raise_when_row_order_ignored(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("lolo", 13, "doctor"),
                ("toto", 12, "cop"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        assert_frame_partially_equals(df1, df2, ["name", "age"], check_row_order=False)

    def test_should_raise_when_row_order_not_ignored(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("lolo", 13, "doctor"),
                ("toto", 12, "cop"),
            ],
            columns=["name", "age", "job"],
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_partially_equals(df1, df2, ["name", "age"], check_row_order=True)

    def test_should_not_raise_when_columns_are_not_in_same_order(self):
        # Given
        df1 = pd.DataFrame(
            [
                ("toto", 12, "developer"),
                ("lolo", 13, "photograph"),
            ],
            columns=["name", "age", "job"],
        )
        df2 = pd.DataFrame(
            [
                ("cop", 12, "toto"),
                ("doctor", 13, "lolo"),
            ],
            columns=["job", "age", "name"],
        )

        # When & Then
        assert_frame_partially_equals(df1, df2, ["name", "age"])


class TestAssertFrameEquals:
    def test_should_assert_2_df_are_equals(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )

        # When
        res = assert_frame_equals(df1, df2)

        # Then
        assert res is None

    def test_should_raise_when_2_df_are_not_equals(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "doctor"))
            .build()
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_equals(df1, df2)

    def test_should_raise_when_row_order_not_ignored(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("lolo", 13, "photograph"))
            .with_row(("toto", 12, "developer"))
            .build()
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_equals(df1, df2, check_row_order=True)

    def test_should_not_raise_when_row_order_ignored(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("lolo", 13, "photograph"))
            .with_row(("toto", 12, "developer"))
            .build()
        )

        # When & Then
        assert_frame_equals(df1, df2, check_row_order=False)

    def test_should_raise_when_df_columns_are_different(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job", "city"])
            .with_row(("toto", 12, "developer", "Paris"))
            .with_row(("lolo", 13, "photograph", "Lyon"))
            .build()
        )

        # When & Then
        with pytest.raises(
            AssertionError,
            match=re.escape(
                "Columns are different. left ones are ['age', 'job', 'name'] "
                "and right ones are ['age', 'city', 'job', 'name']"
            ),
        ):
            assert_frame_equals(df1, df2)

    def test_should_not_raise_when_column_type_ignored(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12.0, "developer"))
            .with_row(("lolo", 13.0, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )

        # When & Then
        assert_frame_equals(df1, df2, check_dtype=False)

    def test_should_raise_when_column_type_not_ignored(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12.0, "developer"))
            .with_row(("lolo", 13.0, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_equals(df1, df2, check_dtype=True)

    def test_should_not_raise_when_columns_order_ignored(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["job", "age", "name"])
            .with_row(("developer", 12, "toto"))
            .with_row(("photograph", 13, "lolo"))
            .build()
        )

        # When & Then
        assert_frame_equals(df1, df2, check_columns_order=False)

    def test_should_raise_when_columns_order_not_ignored(self):
        # Given
        df1 = (
            DataFrameBuilder()
            .with_columns(["name", "age", "job"])
            .with_row(("toto", 12, "developer"))
            .with_row(("lolo", 13, "photograph"))
            .build()
        )
        df2 = (
            DataFrameBuilder()
            .with_columns(["job", "age", "name"])
            .with_row(("developer", 12, "toto"))
            .with_row(("photograph", 13, "lolo"))
            .build()
        )

        # When & Then
        with pytest.raises(
            AssertionError,
        ):
            assert_frame_equals(df1, df2, check_columns_order=True)
```


Environment variables:

- GOPATH=/usr/local/go
- PATH includes $GOPATH/bin

#### Code Samples (Go)

```go
package main

import "fmt"

func SplitDocuments(splitter textsplitter.TextSplitter, docs []*ai.Document) ([]*ai.Document, error) {
	preparedDocs := make([]*ai.Document, 0)
	for _, doc := range docs {
		text := pkg.ContentToText(doc.Content)
		chunks, err := splitter.SplitText(text)
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			if len(chunk) == 0 || OnlyContainsHeaders(chunk) {
				continue
			}
			chunk = strings.TrimSpace(chunk)
			preparedDocs = append(preparedDocs, ai.DocumentFromText(chunk, doc.Metadata))
		}
	}
	return preparedDocs, nil
}
```