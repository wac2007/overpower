package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func isInStringSlice(target string, list []string) bool {
	for _, t := range list {
		if t == target {
			return true
		}
	}
	return false
}

var cluster string

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "Check status of ecs tasks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]
		ctx, cancel := context.WithCancel((context.Background()))
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		defer func() {
			signal.Stop((c))
			cancel()
		}()

		go func() {
			select {
			case <-c:
				cancel()
			case <-ctx.Done():
			}
		}()

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		svc := ecs.New(sess)

		// serviceName := "sherlock-buscape-v3-production"
		arns, err := svc.ListTasks(&ecs.ListTasksInput{
			Cluster:     &cluster,
			ServiceName: &serviceName,
		})
		if err != nil {
			log.Fatal(err)
		}

		tasks, err := svc.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: &cluster,
			Tasks:   arns.TaskArns,
		})
		if err != nil {
			log.Fatal(err)
		}

		r := regexp.MustCompile(`.+\/(.+)$`)

		var (
			rows               []table.Row
			taskDefinitionList []string
		)

		for i, task := range tasks.Tasks {
			arn := r.FindStringSubmatch(*task.TaskArn)[1]
			taskDefinition := r.FindStringSubmatch(*task.TaskDefinitionArn)[1]
			rows = append(rows, table.Row{
				i,
				arn,
				taskDefinition,
				*task.LastStatus,
				*task.DesiredStatus,
			})

			// Check if task definition is in array
			if !isInStringSlice(*task.TaskDefinitionArn, taskDefinitionList) {
				taskDefinitionList = append(taskDefinitionList, *task.TaskDefinitionArn)
			}
		}

		for _, tdArn := range taskDefinitionList {
			tdef, err := svc.DescribeTaskDefinitionWithContext(
				ctx,
				&ecs.DescribeTaskDefinitionInput{
					TaskDefinition: &tdArn,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Print(tdef)
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "Task", "Task Definition", "Last Status", "Desired Status"})
		t.AppendRows(rows)
		t.SetStyle(table.StyleDouble)

		t.SetRowPainter(table.RowPainter(func(row table.Row) text.Colors {
			if row[3] != row[4] {
				return text.Colors{
					text.FgYellow,
				}
			}
			return nil
		}))
		t.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ecsCmd)
	ecsCmd.Flags().StringVarP(&cluster, "cluster", "c", "", "Cluster where service is hosted")
}
