package aws

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSBudgetsBudget_basic(t *testing.T) {
	name := fmt.Sprintf("test-budget-%d", acctest.RandInt())
	configBasicDefaults := newBudgetTestConfigDefaults(name)
	configBasicDefaults.AccountID = "012345678910"
	configBasicUpdate := newBudgetTestConfigUpdate(name)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testCheckBudgetDestroy(name, testAccProvider)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_BasicDefaults(configBasicDefaults),
				Check: resource.ComposeTestCheckFunc(
					testBudgetExists(configBasicDefaults, testAccProvider),
					resource.TestCheckNoResourceAttr("aws_budgets_budget.foo", "account_id"),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name", regexp.MustCompile(configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", configBasicDefaults.LimitAmount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", configBasicDefaults.LimitUnit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicDefaults.TimePeriodStart),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicDefaults.TimePeriodEnd),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", configBasicDefaults.TimeUnit),
				),
			},
			{
				PlanOnly:    true,
				Config:      testAccAWSBudgetsBudgetConfig_WithAccountID(configBasicDefaults),
				ExpectError: regexp.MustCompile("account_id.*" + configBasicDefaults.AccountID),
			},
			{
				Config: testAccAWSBudgetsBudgetConfig_Basic(configBasicUpdate),
				Check: resource.ComposeTestCheckFunc(
					testBudgetExists(configBasicUpdate, testAccProvider),
					resource.TestCheckNoResourceAttr("aws_budgets_budget.foo", "account_id"),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name", regexp.MustCompile(configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", configBasicUpdate.LimitAmount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", configBasicUpdate.LimitUnit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicUpdate.TimePeriodStart),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicUpdate.TimePeriodEnd),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", configBasicUpdate.TimeUnit),
				),
			},
		},
	})
}

func TestAccAWSBudgetsBudget_prefix(t *testing.T) {
	name := "test-budget-"
	configBasicDefaults := newBudgetTestConfigDefaults(name)
	configBasicUpdate := newBudgetTestConfigUpdate(name)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testCheckBudgetDestroy(name, testAccProvider)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_PrefixDefaults(configBasicDefaults),
				Check: resource.ComposeTestCheckFunc(
					testBudgetExists(configBasicDefaults, testAccProvider),
					resource.TestCheckNoResourceAttr("aws_budgets_budget.foo", "account_id"),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name_prefix", regexp.MustCompile(configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", configBasicDefaults.LimitAmount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", configBasicDefaults.LimitUnit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicDefaults.TimePeriodStart),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicDefaults.TimePeriodEnd),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", configBasicDefaults.TimeUnit),
				),
			},

			{
				Config: testAccAWSBudgetsBudgetConfig_Prefix(configBasicUpdate),
				Check: resource.ComposeTestCheckFunc(
					testBudgetExists(configBasicUpdate, testAccProvider),
					resource.TestCheckNoResourceAttr("aws_budgets_budget.foo", "account_id"),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name_prefix", regexp.MustCompile(configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", configBasicUpdate.LimitAmount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", configBasicUpdate.LimitUnit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicUpdate.TimePeriodStart),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicUpdate.TimePeriodEnd),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", configBasicUpdate.TimeUnit),
				),
			},
		},
	})
}

func testBudgetExists(config budgetTestConfig, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["aws_budgets_budget.foo"]
		if !ok {
			return fmt.Errorf("Not found: %s", "aws_budgets_budget.foo")
		}

		accountID, budgetName, err := decodeBudgetsBudgetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed decoding ID: %v", err)
		}

		client := provider.Meta().(*AWSClient).budgetconn
		b, err := client.DescribeBudget(&budgets.DescribeBudgetInput{
			BudgetName: &budgetName,
			AccountId:  &accountID,
		})

		if err != nil {
			return fmt.Errorf("Describebudget error: %v", err)
		}

		if b.Budget == nil {
			return fmt.Errorf("No budget returned %v in %v", b.Budget, b)
		}

		if *b.Budget.BudgetLimit.Amount != config.LimitAmount {
			return fmt.Errorf("budget limit incorrectly set %v != %v", config.LimitAmount, *b.Budget.BudgetLimit.Amount)
		}

		if err := checkBudgetCostTypes(config, *b.Budget.CostTypes); err != nil {
			return err
		}

		if err := checkBudgetTimePeriod(config, *b.Budget.TimePeriod); err != nil {
			return err
		}

		if v, ok := b.Budget.CostFilters[config.FilterKey]; !ok || *v[len(v)-1] != config.FilterValue {
			return fmt.Errorf("cost filter not set properly: %v", b.Budget.CostFilters)
		}

		return nil
	}
}

func checkBudgetTimePeriod(config budgetTestConfig, timePeriod budgets.TimePeriod) error {
	if end, _ := time.Parse("2006-01-02_15:04", config.TimePeriodEnd); *timePeriod.End != end {
		return fmt.Errorf("TimePeriodEnd not set properly '%v' should be '%v'", *timePeriod.End, end)
	}

	if start, _ := time.Parse("2006-01-02_15:04", config.TimePeriodStart); *timePeriod.Start != start {
		return fmt.Errorf("TimePeriodStart not set properly '%v' should be '%v'", *timePeriod.Start, start)
	}

	return nil
}

func checkBudgetCostTypes(config budgetTestConfig, costTypes budgets.CostTypes) error {
	if strconv.FormatBool(*costTypes.IncludeCredit) != config.IncludeCredit {
		return fmt.Errorf("IncludeCredit not set properly '%v' should be '%v'", *costTypes.IncludeCredit, config.IncludeCredit)
	}

	if strconv.FormatBool(*costTypes.IncludeOtherSubscription) != config.IncludeOtherSubscription {
		return fmt.Errorf("IncludeOtherSubscription not set properly '%v' should be '%v'", *costTypes.IncludeOtherSubscription, config.IncludeOtherSubscription)
	}

	if strconv.FormatBool(*costTypes.IncludeRecurring) != config.IncludeRecurring {
		return fmt.Errorf("IncludeRecurring not set properly  '%v' should be '%v'", *costTypes.IncludeRecurring, config.IncludeRecurring)
	}

	if strconv.FormatBool(*costTypes.IncludeRefund) != config.IncludeRefund {
		return fmt.Errorf("IncludeRefund not set properly '%v' should be '%v'", *costTypes.IncludeRefund, config.IncludeRefund)
	}

	if strconv.FormatBool(*costTypes.IncludeSubscription) != config.IncludeSubscription {
		return fmt.Errorf("IncludeSubscription not set properly '%v' should be '%v'", *costTypes.IncludeSubscription, config.IncludeSubscription)
	}

	if strconv.FormatBool(*costTypes.IncludeSupport) != config.IncludeSupport {
		return fmt.Errorf("IncludeSupport not set properly '%v' should be '%v'", *costTypes.IncludeSupport, config.IncludeSupport)
	}

	if strconv.FormatBool(*costTypes.IncludeTax) != config.IncludeTax {
		return fmt.Errorf("IncludeTax not set properly '%v' should be '%v'", *costTypes.IncludeTax, config.IncludeTax)
	}

	if strconv.FormatBool(*costTypes.IncludeUpfront) != config.IncludeUpfront {
		return fmt.Errorf("IncludeUpfront not set properly '%v' should be '%v'", *costTypes.IncludeUpfront, config.IncludeUpfront)
	}

	if strconv.FormatBool(*costTypes.UseBlended) != config.UseBlended {
		return fmt.Errorf("UseBlended not set properly '%v' should be '%v'", *costTypes.UseBlended, config.UseBlended)
	}

	return nil
}

func testCheckBudgetDestroy(budgetName string, provider *schema.Provider) error {
	meta := provider.Meta()
	client := meta.(*AWSClient).budgetconn
	accountID := meta.(*AWSClient).accountid
	_, err := client.DescribeBudget(&budgets.DescribeBudgetInput{
		BudgetName: &budgetName,
		AccountId:  &accountID,
	})
	if !isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
		return fmt.Errorf("Budget '%s' was not deleted properly", budgetName)
	}

	return nil
}

type budgetTestConfig struct {
	AccountID                string
	BudgetName               string
	BudgetType               string
	LimitAmount              string
	LimitUnit                string
	IncludeCredit            string
	IncludeOtherSubscription string
	IncludeRecurring         string
	IncludeRefund            string
	IncludeSubscription      string
	IncludeSupport           string
	IncludeTax               string
	IncludeUpfront           string
	UseBlended               string
	TimeUnit                 string
	TimePeriodStart          string
	TimePeriodEnd            string
	FilterKey                string
	FilterValue              string
}

func newBudgetTestConfigUpdate(name string) budgetTestConfig {
	dateNow := time.Now()
	futureDate := dateNow.AddDate(0, 0, 14)
	return budgetTestConfig{
		BudgetName:               name,
		BudgetType:               "COST",
		LimitAmount:              "500",
		LimitUnit:                "USD",
		FilterKey:                "AZ",
		FilterValue:              "us-east-2",
		IncludeCredit:            "true",
		IncludeOtherSubscription: "true",
		IncludeRecurring:         "true",
		IncludeRefund:            "true",
		IncludeSubscription:      "false",
		IncludeSupport:           "true",
		IncludeTax:               "false",
		IncludeUpfront:           "true",
		UseBlended:               "true",
		TimeUnit:                 "MONTHLY",
		TimePeriodStart:          "2017-01-01_12:00",
		TimePeriodEnd:            futureDate.Format("2006-01-02_15:04"),
	}
}

func newBudgetTestConfigDefaults(name string) budgetTestConfig {
	return budgetTestConfig{
		BudgetName:               name,
		BudgetType:               "COST",
		LimitAmount:              "100",
		LimitUnit:                "USD",
		FilterKey:                "AZ",
		FilterValue:              "us-east-1",
		IncludeCredit:            "true",
		IncludeOtherSubscription: "true",
		IncludeRecurring:         "true",
		IncludeRefund:            "true",
		IncludeSubscription:      "true",
		IncludeSupport:           "true",
		IncludeTax:               "true",
		IncludeUpfront:           "true",
		UseBlended:               "false",
		TimeUnit:                 "MONTHLY",
		TimePeriodStart:          "2017-01-01_12:00",
		TimePeriodEnd:            "2087-06-15_00:00",
	}
}

func testAccAWSBudgetsBudgetConfig_WithAccountID(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	account_id = "{{.AccountID}}"
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	time_period_start = "{{.TimePeriodStart}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_PrefixDefaults(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	time_period_start = "{{.TimePeriodStart}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_Prefix(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	cost_types = {
		include_tax = "{{.IncludeTax}}"
		include_subscription = "{{.IncludeSubscription}}"
		use_blended = "{{.UseBlended}}"
	}
	time_period_start = "{{.TimePeriodStart}}" 
	time_period_end = "{{.TimePeriodEnd}}"
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_BasicDefaults(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	time_period_start = "{{.TimePeriodStart}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_Basic(budgetConfig budgetTestConfig) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.LimitAmount}}"
 	limit_unit = "{{.LimitUnit}}"
	cost_types = {
		include_tax = "{{.IncludeTax}}"
		include_subscription = "{{.IncludeSubscription}}"
		use_blended = "{{.UseBlended}}"
	}
	time_period_start = "{{.TimePeriodStart}}" 
	time_period_end = "{{.TimePeriodEnd}}"
 	time_unit = "{{.TimeUnit}}"
	cost_filters {
		{{.FilterKey}} = "{{.FilterValue}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}
