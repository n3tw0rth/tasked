use crate::args::{Args, AuthOption, Command, ListOption};
use crate::auth::google::GoogleOAuth;
use crate::cli::inline::Inline;
use crate::common::tasks::{GoogleTasks, Tasks, TasksLists};

use anyhow::Result;
use ratatui::style::{Color, Style};
use ratatui::text::{Line, Span, Text};
use ratatui::widgets::Paragraph;

pub struct Cli {
    tasks: GoogleTasks,
}

impl Cli {
    pub async fn new() -> Self {
        Self {
            tasks: GoogleTasks::new().await,
        }
    }

    pub async fn handle_commands(&mut self, args: Args) -> Result<()> {
        let mut google_auth = GoogleOAuth::new().await;
        match args.command {
            Some(Command::Auth { auth }) => match auth {
                AuthOption::Login => google_auth.sign_in().await?,
                AuthOption::Refresh => google_auth.refresh_token().await?,
                AuthOption::Logout => unimplemented!(),
            },
            Some(Command::List { list }) => match list {
                ListOption::Lists => {
                    let tasks = self.tasks.get_tasks_lists().await?;
                    self.cli_get_task_lists(tasks).await?
                }
                ListOption::Tasks => {
                    let tasks_list = self
                        .tasks
                        .get_tasks("MDYyMzk3MDkxNjkyNDIyNzU5MDE6MDow")
                        .await?;
                    self.cli_get_tasks(tasks_list).await?;
                }
            },
            Some(Command::Add { value }) => {
                println!("Adding {}", value);
                self.tasks.add_tasks().await?;
            }
            None => {
                println!("No command provided.");
            }
        }

        Ok(())
    }

    // get_tasks and get_tasks_lists return different results
    // in that case cannot use a match expression right away
    async fn cli_get_task_lists(&mut self, task_lists: TasksLists) -> Result<()> {
        Inline::new().show(|frame| {
            let area = frame.area();
            let lines: Vec<Line<'_>> = task_lists
                .items
                .unwrap()
                .iter()
                .enumerate()
                .map(|(index, item)| {
                    Line::from(vec![Span::styled(
                        format!("{}. {}", index + 1, item.title.clone()),
                        Style::default().fg(Color::Green),
                    )])
                })
                .collect();

            let paragraph = Paragraph::new(lines);
            frame.render_widget(paragraph, area);
        })?;
        Ok(())
    }

    async fn cli_get_tasks(&self, task_lists: TasksLists) -> Result<()> {
        Inline::new().show(|frame| {
            let area = frame.area();
            let lines: Vec<Line<'_>> = task_lists
                .items
                .unwrap()
                .iter()
                .enumerate()
                .map(|(index, item)| {
                    let list_title_line = Line::from(vec![Span::styled(
                        format!("{}. {}", index + 1, item.title.clone()),
                        Style::default().fg(Color::Green),
                    )]);

                    //let task_lines = item.tasks.as_ref().unwrap().iter().map(|t| {
                    //    Line::from(vec![Span::styled(
                    //        format!("- {}", t.title),
                    //        Style::default().fg(Color::Green),
                    //    )])
                    //});

                    Line::from(list_title_line)
                })
                .collect();

            let paragraph = Paragraph::new(lines);
            frame.render_widget(paragraph, area);
        })?;
        Ok(())
    }
}
