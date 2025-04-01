use anyhow::Result;
use chrono::{DateTime, Local};
use regex::{Match, Regex};
use reqwest::{Client, StatusCode};
use serde::Deserialize;
use tracing::{error, info};

use crate::auth::google::GoogleOAuth;
use crate::common::Priority;

const TASKS_API_ENDPOINT: &str = "https://tasks.googleapis.com/tasks/v1";

#[derive(Clone, Debug, Default, Deserialize)]
#[allow(dead_code)]
pub struct TasksLists {
    kind: String,
    etag: String,
    pub items: Option<Vec<TasksList>>,
}

#[derive(Clone, Debug, Deserialize, Default)]
#[allow(dead_code)]
pub struct TasksList {
    kind: String,
    id: String,
    etag: String,
    pub title: String,
    updated: DateTime<Local>,
    pub tasks: Option<Vec<Tasks>>,
}

#[derive(Clone, Debug, Deserialize, Default)]
#[serde(rename_all = "camelCase")]
#[allow(dead_code)]
pub struct Tasks {
    kind: String,
    etag: String,
    pub title: String,
    pub notes: Option<String>,
    updated: String,
    position: String,
    pub status: String,
    web_view_link: String,
    pub priority: Option<Priority>,
}

/// Tasks Lists and Tasks are from two different apis, using this struct to define the additional
/// format for tasks response and update the original TasksList with the items from this struct
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct TasksResponse {
    kind: String,
    etag: String,
    pub items: Option<Vec<Tasks>>,
}

pub struct GoogleTasks {
    pub auth: Box<GoogleOAuth>,
    pub task_lists: TasksLists,
}

impl GoogleTasks {
    pub async fn new() -> Self {
        Self {
            auth: Box::new(GoogleOAuth::new().await),
            task_lists: TasksLists::default(),
        }
    }

    /// If the tasks lists are empty, fetch from the google API
    pub async fn sync(&mut self) -> Result<()> {
        // FIXME: write tasks to a local file
        // At the moment tasks are stored in memory, therefore the tasks should be loaded everytime
        // before running anything else
        if self.task_lists.items.iter().len() <= 0 {
            self.get_tasks_lists().await?;
        }
        Ok(())
    }

    /// Google tasks requires list id to return tasks under a specific list. This function uses
    /// the list id to fetch all the tasks goes under that list id
    pub async fn get_tasks_by_list_id(&mut self, list_id: &str) -> Result<TasksLists> {
        self.sync().await?;

        let url = format!("{}/lists/{}/tasks", TASKS_API_ENDPOINT, list_id);
        let token = self.auth.get_tokens().await.unwrap();

        let tasks = match Client::new()
            .get(url)
            .header("Authorization", format!("Bearer {}", token))
            .send()
            .await
        {
            Err(e) => {
                error!("{}", e);
                Some(vec![Tasks::default()])
            }
            Ok(resp) => match resp.status() {
                StatusCode::OK => Some(resp.json::<TasksResponse>().await.unwrap().items.unwrap()),
                StatusCode::UNAUTHORIZED => {
                    error!("Unauthenicated");
                    self.auth.refresh_token().await?;
                    if let Some(tasklist) =  self.get_tasks_by_list_id(list_id).await.ok();
                }
                _ => {
                    error!("{:?}", resp.json::<TasksLists>().await.ok());
                    Some(vec![Tasks::default()])
                }
            },
        };

        self.update_tasks(list_id, tasks.unwrap()).await?;
        Ok(self.task_lists.clone())
    }

    pub async fn get_all_tasks(&mut self) -> Result<TasksLists> {
        self.sync().await?;
        if let Some(lists) = &self.task_lists.items.clone() {
            for list in lists.iter() {
                let id = list.id.clone();
                self.get_tasks_by_list_id(&id).await?;
            }
        }
        Ok(self.task_lists.clone())
    }

    async fn update_tasks(&mut self, list_id: &str, mut tasks: Vec<Tasks>) -> Result<()> {
        // extract the priority from the tasks
        let tasks: Vec<Tasks> = tasks
            .iter_mut()
            .map(move |task| {
                if let Some(p) = Regex::new(r"\[p\d+\]").unwrap().find(task.title.as_str()) {
                    task.priority = Some(Priority::find(p.as_str()))
                }
                task.clone()
            })
            .collect();

        if let Some(items) = self.task_lists.items.as_mut() {
            if let Some(task_list) = items.iter_mut().into_iter().find(|t| t.id == list_id) {
                task_list.tasks = Some(tasks)
            }
        }
        Ok(())
    }

    pub async fn get_tasks_lists(&mut self) -> Result<TasksLists> {
        let response = match reqwest::Client::new()
            .get(format!("{}{}", TASKS_API_ENDPOINT, "/users/@me/lists"))
            .header(
                "Authorization",
                format!("Bearer {}", self.auth.get_tokens().await.unwrap()),
            )
            .send()
            .await
        {
            Err(e) => {
                error!("{}", e);
                TasksLists::default()
            }
            Ok(resp) => match resp.status() {
                StatusCode::OK => resp.json::<TasksLists>().await.unwrap(),
                StatusCode::UNAUTHORIZED => {
                    error!("Unauthenicated");
                    self.auth.refresh_token().await?;
                    let _ = Box::pin(async {
                        self.get_tasks_lists().await.ok();
                    })
                    .await;
                    TasksLists::default()
                }
                _ => {
                    error!("{:?}", resp.json::<TasksLists>().await.ok());
                    TasksLists::default()
                }
            },
        };
        self.task_lists = response.clone();
        Ok(response)
    }

    pub async fn add_tasks(&self) -> Result<()> {
        unimplemented!()
    }
}
