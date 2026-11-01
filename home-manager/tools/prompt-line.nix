{ ... }:
{
  # ホットキーを受けるには常駐している必要がある。ログイン時に起動し、メニューバーの Quit で終了したときは再起動しない。
  # アプリは /Applications に置く。Accessibility の許可は署名とバンドル ID に紐づく
  launchd.agents.prompt-line = {
    enable = true;
    config = {
      Label = "com.sunagawasei.prompt-line";
      ProgramArguments = [ "/Applications/Prompt Line.app/Contents/MacOS/PromptLine" ];
      RunAtLoad = true;
      KeepAlive = {
        SuccessfulExit = false;
        Crashed = true;
      };
      ProcessType = "Interactive";
      LimitLoadToSessionType = "Aqua";
    };
  };
}
