package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

type Task struct {
	ID   int       `json:"id"`
	Name string    `json:"name"`
	Done bool      `json:"done"`
	Due  time.Time `json:"due"`
}

const filename = "tasks.json"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		handleAdd()
	case "list":
		handleList()
	case "done":
		handleDone()
	case "delete":
		handleDelete()
	case "modify":
		handleModify()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: todo <command> [options]
Commands:
  add     - 添加新任务
  list    - 列出所有任务
  done    - 标记任务完成
  delete  - 删除任务
  modify  - 修改任务

Use 'todo <command> -help' for command details`)
}

func handleAdd() {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "任务内容（必填）")
	due := fs.String("due", "", "截止日期（YYYY-MM-DD）")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		return
	}

	if *name == "" {
		fmt.Println("错误：必须提供任务名称（-name）")
		os.Exit(1)
	}

	var dueTime time.Time
	if *due != "" {
		t, err := time.Parse("2006-01-02", *due)
		if err != nil {
			fmt.Printf("错误：无效的日期格式 - %v\n", err)
			os.Exit(1)
		}
		dueTime = t.Add(23*time.Hour + 59*time.Minute) // 当天23:59
	}

	tasks, _ := readTasks()
	newID := 1
	if len(tasks) > 0 {
		newID = tasks[len(tasks)-1].ID + 1
	}

	tasks = append(tasks, Task{
		ID:   newID,
		Name: *name,
		Due:  dueTime,
	})

	if err := saveTasks(tasks); err != nil {
		fmt.Printf("保存失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Println("任务添加成功！")
}

func handleList() {
	tasks, err := readTasks()
	if err != nil {
		fmt.Printf("读取失败：%v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("当前没有任务")
		return
	}

	now := time.Now()
	fmt.Println("任务列表：")
	for _, t := range tasks {
		status := "✓"
		if !t.Done {
			status = " "
		}

		dueInfo := "无截止日期"
		if !t.Due.IsZero() {
			dueInfo = t.Due.Format("2006-01-02")
			if t.Due.Before(now) && !t.Done {
				dueInfo += " (已过期)"
			} else if days := int(t.Due.Sub(now).Hours()/24) + 1; days > 0 && !t.Done {
				dueInfo += fmt.Sprintf(" (%d天后到期)", days)
			}
		}

		fmt.Printf("[%s] %d. %-20s %s\n", status, t.ID, t.Name, dueInfo)
	}
}

func handleDone() {
	fs := flag.NewFlagSet("done", flag.ExitOnError)
	id := fs.Int("id", 0, "要完成的任务ID")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		return
	}

	if *id == 0 {
		fmt.Println("错误：必须提供任务ID（-id）")
		os.Exit(1)
	}

	tasks, err := readTasks()
	if err != nil {
		fmt.Printf("读取失败：%v\n", err)
		os.Exit(1)
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == *id {
			tasks[i].Done = true
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("错误：找不到ID为%d的任务\n", *id)
		os.Exit(1)
	}

	if err := saveTasks(tasks); err != nil {
		fmt.Printf("保存失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Println("任务标记为已完成！")
}

func handleDelete() {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	id := fs.Int("id", 0, "要删除的任务ID")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		return
	}

	if *id == 0 {
		fmt.Println("错误：必须提供任务ID（-id）")
		os.Exit(1)
	}

	tasks, err := readTasks()
	if err != nil {
		fmt.Printf("读取失败：%v\n", err)
		os.Exit(1)
	}

	newTasks := make([]Task, 0, len(tasks))
	found := false
	for _, t := range tasks {
		if t.ID == *id {
			found = true
		} else {
			newTasks = append(newTasks, t)
		}
	}

	if !found {
		fmt.Printf("错误：找不到ID为%d的任务\n", *id)
		os.Exit(1)
	}

	if err := saveTasks(newTasks); err != nil {
		fmt.Printf("保存失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Println("任务删除成功！")
}

func handleModify() {
	fs := flag.NewFlagSet("modify", flag.ExitOnError)
	id := fs.Int("id", 0, "要修改的任务ID")
	name := fs.String("name", "", "新的任务内容")
	due := fs.String("due", "", "新的截止日期（YYYY-MM-DD）")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		return
	}

	if *id == 0 {
		fmt.Println("错误：必须提供任务ID（-id）")
		os.Exit(1)
	}

	tasks, err := readTasks()
	if err != nil {
		fmt.Printf("读取失败：%v\n", err)
		os.Exit(1)
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == *id {
			found = true
			if *name != "" {
				tasks[i].Name = *name
			}
			if *due != "" {
				t, err := time.Parse("2006-01-02", *due)
				if err != nil {
					fmt.Printf("错误：无效的日期格式 - %v\n", err)
					os.Exit(1)
				}
				tasks[i].Due = t.Add(23 * time.Hour)
			}
			break
		}
	}

	if !found {
		fmt.Printf("错误：找不到ID为%d的任务\n", *id)
		os.Exit(1)
	}

	if err := saveTasks(tasks); err != nil {
		fmt.Printf("保存失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Println("任务修改成功！")
}

func readTasks() ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var tasks []Task
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func saveTasks(tasks []Task) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}
