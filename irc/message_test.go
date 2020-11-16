package irc

import "testing"

func TestIRCv3(t *testing.T) {
	messages := []string{
		"@msg-id=slow_off :tmi.twitch.tv NOTICE #channel :This room is no longer in slow mode.",
		"@color=#0D4200;display-name=TWITCH_UserNaME;emotes=25:0-4,12-16/1902:6-10;subscriber=0;turbo=1;user-type=global_mod :twitch_username!twitch_username@twitch_username.tmi.twitch.tv PRIVMSG #channel :Kappa Keepo Kappa",
		"@badge-info=;badges=;color=#1E90FF;display-name=cbdg;emote-only=1;emotes=300949179:0-9;flags=;id=477f0595-3183-4520-85d7-123e2018806b;mod=0;msg-id=highlighted-message;room-id=24761645;subscriber=0;tmi-sent-ts=1578179988796;turbo=0;user-id=53381086;user-type= :cbdg!cbdg@cbdg.tmi.twitch.tv PRIVMSG #cirno_tv :naroStaryn",
		"@badge-info=subscriber/68;badges=subscriber/60;color=#6B00B8;display-name=MrXtacle;emotes=;flags=;id=16894214-c514-4612-a979-de9280ca437a;login=mrxtacle;mod=0;msg-id=resub;msg-param-cumulative-months=68;msg-param-months=0;msg-param-should-share-streak=0;msg-param-sub-plan-name=Baka\\sBrigade!;msg-param-sub-plan=1000;room-id=24761645;subscriber=1;system-msg=MrXtacle\\ssubscribed\\sat\\sTier\\s1.\\sThey've\\ssubscribed\\sfor\\s68\\smonths!;tmi-sent-ts=1578181991964;user-id=28413930;user-type= :tmi.twitch.tv USERNOTICE #cirno_tv :Look at me, I'm all gwown up RainbowDaijoubu",
		"@color=#0D4200;display-name=TWITCH_UserNaME;emotes=25:0-4,12-16/1902:6-10;subscriber=0;turbo=1;user-type=global_mod :twitch_username!twitch_username@twitch_username.tmi.twitch.tv PRIVMSG #channel Kappa",
	}
	payloads := []string{
		"This room is no longer in slow mode.",
		"Kappa Keepo Kappa",
		"naroStaryn",
		"Look at me, I'm all gwown up RainbowDaijoubu",
		"Kappa",
	}

	for i, m := range messages {
		msg := ParseMessage(m)
		t.Logf("Message: %#v", msg)
		if p := msg.Trailer(1); p != payloads[i] {
			t.Errorf("Wrong payload: %q != %q", p, payloads[i])
		}
	}
}

// benchMessages contains some complex messages for "worst-case" benchmarking
var benchMessages = []string{
	"@badge-info=;badges=glhf-pledge/1;color=#5F9EA0;display-name=ThePositiveBot;emotes=;flags=;id=fa6644c8-3474-4d57-8327-e1c3ef8d2b21;mod=0;room-id=22484632;subscriber=0;tmi-sent-ts=1604966456146;turbo=0;user-id=425363834;user-type= :thepositivebot!thepositivebot@thepositivebot.tmi.twitch.tv PRIVMSG #forsen :\u0001ACTION [Cookies] [Masters] wisehardo, you have 162 cookies! You can also claim your next cookie now by doing ?cookie!\u0001",
	"@badge-info=subscriber/1;badges=subscriber/0;color=#FF0000;display-name=marteenstemat;emotes=;flags=;id=e94972ac-e1ae-4182-9710-360bedadd2f8;mod=0;room-id=22484632;subscriber=1;tmi-sent-ts=1604966549514;turbo=0;user-id=594963954;user-type= :marteenstemat!marteenstemat@marteenstemat.tmi.twitch.tv PRIVMSG #forsen :@gyoubu_masataka_oniwaaaaa you need to watch vod that forsen talking about his \"job\" and how it's work pepeLaugh",
	"@badge-info=subscriber/75;badges=moderator/1,subscriber/72;color=#12AFED;display-name=Snusbot;emotes=173378:264-273/684688:331-339/36391:47-55/60257:66-72/96553:153-162/115996:164-170/116051:172-178/173372:256-262/1558721:427-433/555437:321-329/31097:27-36/36535:57-64/89640:94-103/89650:119-127/116053:188-194/118074:236-242/1558719:416-425/31021:19-25/60391:74-83/89678:129-139/184115:293-299/521050:313-319/90377:141-151/116245:204-212/175766:275-283/1171397:374-380/1271995:382-390/67683:85-92/116256:214-222/239535:301-311/1565952:461-473/1565929:447-459/116052:180-186/116055:196-202/116273:224-234/122261:244-254/684692:341-347/1361610:392-402/1558723:435-445/1565958:475-488/31100:38-45/89641:105-117/177866:285-291/696755:349-362/780629:364-372/1479466:404-414;flags=;id=f816c0db-d315-4609-ad9d-f0c47c3eead2;mod=1;room-id=22484632;subscriber=1;tmi-sent-ts=1604967031958;turbo=0;user-id=62541963;user-type=mod :snusbot!snusbot@snusbot.tmi.twitch.tv PRIVMSG #forsen :Subscriber emotes: forsenW forsenBoys forsenRP forsenDDK forsenSS forsenX forsenPuke forsenIQ forsenWhip forsenSleeper forsenGun forsenPuke2 forsenKnife forsenLewd forsenH forsen1 forsen2 forsen3 forsen4 forsenLUL forsenDED forsenFeels forsenO forsenPuke3 forsenY forsenGASM forsenWut forsenS forsenT forsenThink forsenE forsenBee forsenKek forsenL forsenRedSonic forsenWTF forsenK forsenDab forsenPuke5 forsenWeird forsenTake forsenA forsenBreak forsenLicence forsenPosture forsenPosture1",
	"@badge-info=;badges=glhf-pledge/1;color=#1E90FF;display-name=radekdarade;emotes=;flags=;id=28f21eab-53f3-4ccd-aeee-c43dcad491a1;login=radekdarade;mod=0;msg-id=raid;msg-param-displayName=radekdarade;msg-param-login=radekdarade;msg-param-profileImageURL=https://static-cdn.jtvnw.net/jtv_user_pictures/70f6d58c-1f74-47c7-96bb-7e2e931c0e36-profile_image-70x70.png;msg-param-viewerCount=1;room-id=22484632;subscriber=0;system-msg=1\\sraiders\\sfrom\\sradekdarade\\shave\\sjoined!;tmi-sent-ts=1605002353635;user-id=208475120;user-type= :tmi.twitch.tv USERNOTICE #forsen",
}

var result *Message

func BenchmarkParser(b *testing.B) {
	var r *Message
	for n := 0; n < b.N; n++ {
		r = ParseMessage(benchMessages[n&3])
	}
	result = r
}
