package cli_ui

func ExampleShowRows() {

	tableHeader := []string{
		"id", "name", "age",
	}
	tableBody := [][]string{
		{
			"1", "Tom", "18",
		},
		{
			"2", "Ada", "26",
		},
		{
			"3", "Sam", "30",
		},
	}
	tableFooter := []string{
		"footer1", "footer2", "footer",
	}
	setBorder := true
	ShowRows(tableHeader, tableBody, tableFooter, setBorder)
	// Output:
	// *********** Row 0 **********
	//
	//	  id:	1
	//	name:	Tom
	//	 age:	18
	//
	//*********** Row 1 **********
	//
	//	  id:	2
	//	name:	Ada
	//	 age:	26
	//
	//*********** Row 2 **********
	//
	//	  id:	3
	//	name:	Sam
	//	 age:	30

}

func ExampleShowTable() {

	tableHeader := []string{
		"id", "name", "age",
	}
	tableBody := [][]string{
		{
			"1", "Tom", "18",
		},
		{
			"2", "Ada", "26",
		},
		{
			"3", "Sam", "30",
		},
	}
	tableFooter := []string{
		"footer1", "footer2", "footer",
	}
	setBorder := true
	ShowTable(tableHeader, tableBody, tableFooter, setBorder)

	// Output:
	// +---------+---------+--------+
	// |   ID    |  NAME   |  AGE   |
	// +---------+---------+--------+
	// |       1 | Tom     |     18 |
	// |       2 | Ada     |     26 |
	// |       3 | Sam     |     30 |
	// +---------+---------+--------+
	// | FOOTER1 | FOOTER2 | FOOTER |
	// +---------+---------+--------+
}
