// All Humble Bundle bundle keys have this length
const FULL_KEY_SIZE: usize = 16;

pub struct KeyMatch {
    keys: Vec<String>,
    target: String,
}

impl KeyMatch {
    pub fn new(keys: Vec<String>, target: &str) -> Self {
        Self {
            keys,
            target: target.to_owned(),
        }
    }

    fn is_full_key(key: &str) -> bool {
        key.len() == FULL_KEY_SIZE
    }

    /// Perform a case-insensitive search and find any key that starts with
    /// the given `target` value.
    ///
    /// If `target` is already a full bundle key and not a partial key, then
    /// the `target` will be returned without any search.
    pub fn get_matches(&self) -> Vec<String> {
        if Self::is_full_key(&self.target) {
            return vec![self.target.clone()];
        }

        let lowercase_target = self.target.to_lowercase();
        self.keys
            .iter()
            .filter(|k| k.to_lowercase().starts_with(&lowercase_target))
            .cloned()
            .collect()
    }
}

#[test]
fn test_exact_match() {
    let keys = vec!["1AaAaA".to_owned(), "2BbBbB".to_owned()];
    let target = "1AaAaA".to_owned();

    let key_match = KeyMatch::new(keys, &target);
    assert_eq!(key_match.get_matches(), vec!["1AaAaA".to_owned()]);
}

#[test]
fn test_single_match() {
    let keys = vec!["1aAa".to_owned(), "2bbbb".to_owned()];
    let target = "1a".to_owned();

    let key_match = KeyMatch::new(keys, &target);
    assert_eq!(key_match.get_matches(), vec!["1aAa".to_owned()]);
}

#[test]
fn test_no_match() {
    let keys = vec!["1aaa".to_owned(), "2bbbb".to_owned()];
    let target = "3c".to_owned();

    let key_match = KeyMatch::new(keys, &target);
    assert!(key_match.get_matches().is_empty());
}

#[test]
fn test_multiple_matches() {
    let keys = vec!["1aaa".to_owned(), "1aXXX".to_owned(), "2bbbb".to_owned()];
    let target = "1a".to_owned();

    let key_match = KeyMatch::new(keys, &target);
    assert_eq!(
        key_match.get_matches(),
        vec!["1aaa".to_owned(), "1aXXX".to_owned()]
    );
}
