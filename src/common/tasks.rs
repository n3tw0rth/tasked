use anyhow::Result;
use chrono::{DateTime, Local};
use reqwest::StatusCode;
use serde::Deserialize;
use tracing::error;

use crate::auth::google::GoogleOAuth;

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
    notes: Option<String>,
    updated: String,
    position: String,
    status: String,
    web_view_link: String,
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

    pub async fn get_tasks(&mut self, list_id: &str) -> Result<TasksLists> {
        // populate the tasks lists if the data is not already fetched
        if self.task_lists.items.iter().len() <= 0 {
            self.get_tasks_lists().await?;
        }

        let tasks = match reqwest::Client::new()
            .get(format!(
                "{}{}",
                TASKS_API_ENDPOINT,
                format!("/lists/{}/tasks", list_id)
            ))
            .header(
                "Authorization",
                format!("Bearer {}", self.auth.get_tokens().await.unwrap()),
            )
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
                    let _ = Box::pin(async {
                        self.get_tasks_lists().await.ok();
                    })
                    .await;
                    Some(vec![Tasks::default()])
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

    async fn update_tasks(&mut self, list_id: &str, tasks: Vec<Tasks>) -> Result<()> {
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
