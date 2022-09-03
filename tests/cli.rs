use assert_cmd::Command;

#[test]
fn runs() {
    let mut cmd = Command::cargo_bin("humble-cli").unwrap();
    let output = cmd.output().unwrap();
    assert_eq!(output.status.code(), Some(2));
    let msg = String::from_utf8(output.stderr).expect("failed to convert to String");
    assert!(msg.contains("missing"));
}
