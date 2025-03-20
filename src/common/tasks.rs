use anyhow::Result;
use chrono::{DateTime, Local};
use reqwest::StatusCode;
use serde::Deserialize;
use tracing::error;

use crate::auth::google::GoogleOAuth;

const TASKS_API_ENDPOINT: &str = "https://tasks.googleapis.com/tasks/v1";

#[derive(Clone, Debug, Default, Deserialize)]
pub struct TasksLists {
    kind: String,
    etag: String,
    pub items: Option<Vec<TasksList>>,
}

#[derive(Clone, Debug, Deserialize)]
pub struct TasksList {
    kind: String,
    id: String,
    etag: String,
    pub title: String,
    updated: DateTime<Local>,
    tasks: Option<Vec<Tasks>>,
}

#[derive(Clone, Debug, Deserialize)]
pub struct Tasks {
    kind: String,
    etag: String,
    title: String,
    notes: Option<String>,
    updated: String,
    posistion: i8,
    status: String,
    web_view_link: String,
}

/// Tasks Lists and Tasks are from two different apis, using this struct to define the additional
/// format for tasks response and update the original TasksList with the items from this struct
#[derive(Debug, Deserialize)]
pub struct ArbitaryTaskList {}

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

    pub async fn get_tasks(&self, _list: &str) -> Result<Tasks> {
        unimplemented!()
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
