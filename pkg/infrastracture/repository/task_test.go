package repository_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/testutil"
)

func Test_TaskRepository_ListTask(t *testing.T) {

	type args struct {
		status *string
		limit  *uint
		offset *uint
	}
	type wants struct {
		tasks []*entities.Task
		error error
	}

	tests := []struct {
		name    string
		prepare func(tx infrastracture.Transactor) ([]*entities.Task, error)
		args    args
		wants   wants
	}{
		{
			name: "全件取得",
			prepare: func(tx infrastracture.Transactor) ([]*entities.Task, error) {
				tasks := make([]*entities.Task, 0, 3)
				for i := 0; i < 3; i++ {
					s := &entities.Task{
						TaskId:     fmt.Sprintf("task_id_%d", i+1),
						TaskStatus: "complete",
						Title:      "title",
						PageUrl:    "page_url",
					}
					tasks = append(tasks, s)
				}
				builder := goqu.
					Insert("tasks").
					Cols("task_id", "task_status", "title", "page_url", "created_at", "updated_at").
					Rows(tasks)
				query, _, err := builder.ToSQL()
				if err != nil {
					return nil, fmt.Errorf("failed to ToSQL: %v", err)
				}
				if _, err := tx.ExecContext(context.Background(), query); err != nil {
					return nil, fmt.Errorf("failed to ExecContext: %v", err)
				}
				return tasks, nil
			},
			args: args{
				status: nil,
				limit:  testutil.UintPtr(10),
				offset: testutil.UintPtr(0),
			},
			wants: wants{
				tasks: func() []*entities.Task {
					tasks := make([]*entities.Task, 0, 3)
					for i := 0; i < 3; i++ {
						s := &entities.Task{
							TaskId:     fmt.Sprintf("task_id_%d", i+1),
							TaskStatus: "complete",
							Title:      "title",
							PageUrl:    "page_url",
						}
						tasks = append(tasks, s)
					}
					sort.Slice(tasks, func(i, j int) bool {
						return tasks[i].TaskId > tasks[j].TaskId
					})
					return tasks
				}(),
				error: nil,
			},
		},
		{
			name: "指定件数取得",
			prepare: func(tx infrastracture.Transactor) ([]*entities.Task, error) {
				tasks := make([]*entities.Task, 0, 10)
				for i := 0; i < 10; i++ {
					s := &entities.Task{
						TaskId:     fmt.Sprintf("task_id_%d", i+1),
						TaskStatus: "complete",
						Title:      "title",
						PageUrl:    "page_url",
					}
					tasks = append(tasks, s)
				}
				builder := goqu.
					Insert("tasks").
					Cols("task_id", "task_status", "title", "page_url", "created_at", "updated_at").
					Rows(tasks)
				query, _, err := builder.ToSQL()
				if err != nil {
					return nil, fmt.Errorf("failed to ToSQL: %v", err)
				}
				if _, err := tx.ExecContext(context.Background(), query); err != nil {
					return nil, fmt.Errorf("failed to ExecContext: %v", err)
				}
				return tasks, nil
			},
			args: args{
				status: nil,
				limit:  testutil.UintPtr(5),
				offset: testutil.UintPtr(1),
			},
			wants: wants{
				tasks: func() []*entities.Task {
					tasks := make([]*entities.Task, 0, 5)
					tasks = append(tasks, &entities.Task{TaskId: "task_id_9", TaskStatus: "complete", Title: "title", PageUrl: "page_url"})
					tasks = append(tasks, &entities.Task{TaskId: "task_id_8", TaskStatus: "complete", Title: "title", PageUrl: "page_url"})
					tasks = append(tasks, &entities.Task{TaskId: "task_id_7", TaskStatus: "complete", Title: "title", PageUrl: "page_url"})
					tasks = append(tasks, &entities.Task{TaskId: "task_id_6", TaskStatus: "complete", Title: "title", PageUrl: "page_url"})
					tasks = append(tasks, &entities.Task{TaskId: "task_id_5", TaskStatus: "complete", Title: "title", PageUrl: "page_url"})
					return tasks
				}(),
				error: nil,
			},
		},
		{
			name: "ステータスrequest取得",
			prepare: func(tx infrastracture.Transactor) ([]*entities.Task, error) {
				tasks := make([]*entities.Task, 0, 10)
				for i := 0; i < 10; i++ {
					s := &entities.Task{
						TaskId: fmt.Sprintf("task_id_%d", i+1),
						TaskStatus: func() string {
							if i < 7 {
								return "complete"
							} else {
								return "request"
							}
						}(),
						Title:   "title",
						PageUrl: "page_url",
					}
					tasks = append(tasks, s)
				}
				builder := goqu.
					Insert("tasks").
					Cols("task_id", "task_status", "title", "page_url", "created_at", "updated_at").
					Rows(tasks)
				query, _, err := builder.ToSQL()
				if err != nil {
					return nil, fmt.Errorf("failed to ToSQL: %v", err)
				}
				if _, err := tx.ExecContext(context.Background(), query); err != nil {
					return nil, fmt.Errorf("failed to ExecContext: %v", err)
				}
				return tasks, nil
			},
			args: args{
				status: testutil.StrPtr("request"),
				limit:  testutil.UintPtr(10),
				offset: testutil.UintPtr(0),
			},
			wants: wants{
				tasks: func() []*entities.Task {
					tasks := make([]*entities.Task, 0, 5)
					tasks = append(tasks, &entities.Task{TaskId: "task_id_10", TaskStatus: "request", Title: "title", PageUrl: "page_url"})
					tasks = append(tasks, &entities.Task{TaskId: "task_id_9", TaskStatus: "request", Title: "title", PageUrl: "page_url"})
					tasks = append(tasks, &entities.Task{TaskId: "task_id_8", TaskStatus: "request", Title: "title", PageUrl: "page_url"})
					return tasks
				}(),
				error: nil,
			},
		},
	}

	dbConfig := &config.RDBConfig{RDBDsn: testutil.RDBDNSForTest}
	dbHandler, err := infrastracture.NewDBHandler(dbConfig)
	if err != nil {
		t.Fatalf("failed to NewDBHandler: %v", err)
	}

	repo := repository.NewTaskRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := dbHandler.GetTransaction()
			if err != nil {
				t.Fatalf("failed to GetTransaction: %v", err)
			}
			defer tx.Rollback()

			if tt.prepare != nil {
				_, err := tt.prepare(tx)
				if err != nil {
					t.Fatalf("failed to prepare: %v", err)
				}
			}

			input := &repository.ListTaskInput{
				Status: tt.args.status,
				Limit:  tt.args.limit,
				Offset: tt.args.offset,
			}

			tasks, err := repo.ListTask(context.Background(), tx, input)
			if err != tt.wants.error {
				t.Errorf("got: %v, want: %v", err, tt.wants.error)
			}

			cmpOpts := cmpopts.IgnoreFields(entities.Task{}, "Id", "CreatedAt", "UpdatedAt")
			if diff := cmp.Diff(tasks, tt.wants.tasks, cmpOpts); diff != "" {
				t.Errorf("got: %v, want: %v", tasks, tt.wants.tasks)
				var t entities.Tasks = tasks
				fmt.Println(t.JSON())
				t = tt.wants.tasks
				fmt.Println(t.JSON())
			}
		})
	}
}
