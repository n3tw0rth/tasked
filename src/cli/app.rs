use crate::args::{Args, AuthOption, Command, ListOption};
use crate::auth::google::GoogleOAuth;
use crate::cli::inline::Inline;
use crate::common::tasks::{GoogleTasks, Tasks, TasksList, TasksLists};
use crate::common::Priority;

use ansi_term::Colour;
use anyhow::Result;
use tracing::info;
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
                    let tasks_list = self.tasks.get_all_tasks().await?;
                    self.cli_get_tasks(tasks_list).await?;
                }
            },
            Some(Command::Add { value }) => {
                info!("Adding {}", value);
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
        Inline::new().show(|| {
            task_lists
                .items
                .unwrap()
                .iter()
                .enumerate()
                .for_each(|(index, item)| {
                    println!(
                        "{}. {}",
                        Colour::Cyan.paint(index.saturating_add(1).to_string()),
                        Colour::Blue.paint(item.title.clone())
                    );
                });
        })?;
        Ok(())
    }

    async fn cli_get_tasks(&self, task_lists: TasksLists) -> Result<()> {
        Inline::new().show(|| {
            task_lists
                .items
                .unwrap_or(vec![TasksList::default()])
                .iter()
                .enumerate()
                .for_each(|(index, item)| {
                    println!(
                        "{}. {}",
                        Colour::Cyan.paint(index.saturating_add(1).to_string()),
                        Colour::Blue.paint(item.title.clone())
                    );
                    item.tasks
                        .as_ref()
                        .unwrap_or(&vec![Tasks::default()])
                        .iter()
                        .for_each(|task| {
                            let title = task.title.clone();
                            println!(
                                "  {}",
                                Priority::P1.color(
                                    task.priority
                                        .clone()
                                        .unwrap_or(Priority::default())
                                        .color(title)
                                )
                            )
                        })
                });
        })?;
        Ok(())
    }
}
