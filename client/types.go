package client

type CurrencyFormat struct {
	IsoCode          string `json:"iso_code"`
	ExampleFormat    string `json:"example_format"`
	DecimalDigits    int    `json:"decimal_digits"`
	DecimalSeparator string `json:"decimal_separator"`
	SymbolFirst      bool   `json:"symbol_first"`
	GroupSeparator   string `json:"group_separator"`
	CurrencySymbol   string `json:"currency_symbol"`
	DisplaySymbol    bool   `json:"display_symbol"`
}

type Budget struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	LastModifiedOn string `json:"last_modified_on"`
	FirstMonth     string `json:"first_month"`
	LastMonth      string `json:"last_month"`
	DateFormat     struct {
		Format string `json:"format"`
	} `json:"date_format"`
	CurrencyFormat CurrencyFormat `json:"currency_format"`
}

type Category struct {
	Id                      string `json:"id"`
	CategoryGroupId         string `json:"category_group_id"`
	Name                    string `json:"name"`
	Hidden                  bool   `json:"hidden"`
	OriginalCategoryGroupId string `json:"original_category_group_id"`
	Note                    string `json:"note"`
	Budgeted                int64  `json:"budgeted"`
	Activity                int64  `json:"activity"`
	Balance                 int64  `json:"balance"`
	GoalType                string `json:"goal_type"`
	GoalCreationMonth       string `json:"goal_creation_month"`
	GoalTarget              int64  `json:"goal_target"`
	GoalTargetMonth         string `json:"goal_target_month"`
	GoalPercentageComplete  int64  `json:"goal_percentage_complete"`
	Deleted                 bool   `json:"deleted"`
}

type CategoryGroup struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	Hidden     bool       `json:"hidden"`
	Deleted    bool       `json:"deleted"`
	Categories []Category `json:"categories"`
}
