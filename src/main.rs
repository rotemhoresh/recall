use std::{collections::BTreeMap, env, fs, path::PathBuf};

use anyhow::{Context, anyhow};
use clap::{Parser, Subcommand};

const FILE_NAME: &str = ".recalls";

struct Recalls {
    path: PathBuf,
    cwd: PathBuf,
    recalls: BTreeMap<String, Vec<String>>,
}

impl Recalls {
    pub fn new(path: PathBuf, cwd: PathBuf) -> anyhow::Result<Self> {
        // If there is no file (which is the primary reason of this failing),
        // we want to just treat it as empty recalls map.
        let content = fs::read_to_string(&path);
        let recalls: BTreeMap<String, Vec<String>> = match content {
            Ok(content) => serde_json::from_str(&content)?,
            Err(_) => BTreeMap::new(),
        };

        Ok(Self { path, cwd, recalls })
    }

    pub fn add(&mut self, notes: Vec<String>) -> anyhow::Result<()> {
        let cwd = self
            .cwd
            .to_str()
            .ok_or_else(|| anyhow!("failed to convert cwd to string"))?
            .to_owned();

        self.recalls.entry(cwd).or_default().extend(notes);

        Ok(())
    }

    pub fn remove(&mut self, mut indexes: Vec<usize>) -> anyhow::Result<()> {
        let cwd = self
            .cwd
            .to_str()
            .ok_or_else(|| anyhow!("failed to convert cwd to string"))?;

        if let Some(list) = self.recalls.get_mut(cwd) {
            indexes.dedup();
            indexes.sort();

            // Because we iterate in reverese, the indexes will be removed
            // from the original list, e.g.:
            //
            // Original:  [ 0, 1, 2, 3, 4, 5 ]
            // To remove: [ 0,       3, 4    ]
            // After:     [    1, 2,       5 ]
            for &idx in indexes.iter().rev() {
                if idx < list.len() {
                    list.remove(idx);
                }
            }
        }

        Ok(())
    }

    pub fn get_list(&self) -> anyhow::Result<Option<&Vec<String>>> {
        let cwd = self
            .cwd
            .to_str()
            .ok_or_else(|| anyhow!("failed to convert cwd to string"))?;

        Ok(self.recalls.get(cwd))
    }

    pub fn write(&self) -> anyhow::Result<()> {
        let content = serde_json::to_string(&self.recalls)?;
        fs::write(&self.path, &content).with_context(|| "failed to write recalls to file")
    }
}

/// A tool for recalling where you left off by leaving notes for yourself.
///
/// When run without any subcommand, the recall list associated with the current working directory will be printed.
#[derive(Debug, Parser)]
#[command(name = "recall")]
struct Cli {
    #[command(subcommand)]
    command: Option<Command>,
}

#[derive(Debug, Subcommand)]
enum Command {
    /// Add notes to the recall list associated with the current working directory
    Add { notes: Vec<String> },
    /// Remove notes by index from the recall list associated with the current working directory
    Rm { indexes: Vec<usize> },
}

fn main() -> anyhow::Result<()> {
    let args = Cli::parse();
    let path = dirs::home_dir()
        .ok_or_else(|| anyhow!("failed to get home dir"))
        .map(|mut p| {
            p.push(FILE_NAME);
            p
        })?;
    let cwd = env::current_dir().with_context(|| "failed to get cwd")?;

    let mut recalls = Recalls::new(path, cwd)?;

    match args.command {
        Some(command) => match command {
            Command::Add { notes } => {
                recalls.add(notes)?;
                recalls.write()?;
            }
            Command::Rm { indexes } => {
                recalls.remove(indexes)?;
                recalls.write()?;
            }
        },
        None => {
            if let Some(list) = recalls.get_list()? {
                for (idx, line) in list.iter().enumerate() {
                    println!("[{idx}] {line}");
                }
            }
        }
    }

    Ok(())
}
