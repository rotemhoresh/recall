#![feature(let_chains)]

use std::{collections::BTreeMap, env, fs, path::PathBuf};

use anyhow::{Context, anyhow};
use clap::{Parser, Subcommand};
use regex::{Captures, Regex};
use supports_hyperlinks::Stream;

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
                let re_path = Regex::new(r"`([^`]+)`").unwrap();

                for (idx, line) in list.iter().enumerate() {
                    let line = re_path.replace_all(line, |c: &Captures| {
                        let path = c.get(1).unwrap().as_str();

                        // Since we can't really fail here, we will
                        // just use an the unhyperlinked path if the
                        // hyperlink generation failed.
                        if supports_hyperlinks::on(Stream::Stdout)
                            && let Ok(hyperlink) = hyperlink(path)
                        {
                            hyperlink
                        } else {
                            path.to_owned()
                        }
                    });
                    println!("[{idx}] {line}");
                }
            }
        }
    }

    Ok(())
}

/// Convert a path to a hyperlink, which includes:
///
/// - resolving to an absolute path.
/// - making sure that the path exists.
/// - wrapping in a hyperlink ANSI escape codes.
///
/// Which makes a result that is the original path, which links to the
/// resolved absolute path.
#[allow(clippy::iter_nth_zero)]
fn hyperlink(path: &str) -> anyhow::Result<String> {
    // Inspired by coreutils' `ls` hyperlink feature.
    // https://github.com/coreutils/coreutils/blob/0032e336e50c86186a01dbfa77364bc9c12235c1/src/ls.c#L4774

    // fs::canonicalize can't handle tilde (i.e., converting it to home dir),
    // thus we need to resolve `~`.
    let mut resolved_path = path.to_owned(); // OPTIMIZE: is there a way without cloning?
    // This is necessary so we won't expand a path that starts with a file
    // or a directory that starts with a tilde, as it is a valid character
    // in a file name.
    if resolved_path.starts_with("~/") || resolved_path == "~" {
        resolved_path.replace_range(
            0..1, // OPTIMIZE: is there a better way to do this?
            dirs::home_dir()
                .ok_or_else(|| anyhow!("failed to get home dir"))?
                .to_str()
                .ok_or_else(|| anyhow!("failed to convert cwd to string"))?,
        )
    }

    let h = gethostname::gethostname();
    let hostname = h
        .to_str()
        .ok_or_else(|| anyhow!("failed to convert host name to str"))?;
    let n = fs::canonicalize(resolved_path).with_context(|| "failed to make path absolute")?;
    let absolute_path = n
        .to_str()
        .ok_or_else(|| anyhow!("failed to convert path to str"))?;

    let separator = if absolute_path.chars().nth(0).is_some_and(|c| c != '/') {
        "/"
    } else {
        ""
    };

    Ok(format!(
        "\x1b]8;;file://{}{}{}\x07{}\x1b]8;;\x07",
        hostname, separator, absolute_path, path
    ))
}
