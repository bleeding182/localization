
import Foundation

// Do _not_ modify
// https://docs.google.com/spreadsheets/d/1upHiDHWu5m30tYdhMDP4GXheOWUE4r3VrHfmAUXiuyI#gid=0
// Last updated at 2018-05-02T21:18:12+02:00

// swiftlint:disable line_length
public struct Strings {
    public struct Greeting {
        static let GreetingHelloWorld = Strings.localized("greeting_hello_world", value: "Hello, world!", comment: "Default greeting.")
        static let GreetingText = Strings.localized("greeting_text", value: "Isn't this a nice iOS device?", comment: "Overridden per platform")
    }
    
    public struct SongLine {
        static let SongLineBottlesOfBeerPlOne = Strings.localized("song_line__bottles_of_beer__pl_one", value: "%1$d bottle of beer on the wall, %1$d bottle of beer.")
        static let SongLineBottlesOfBeerPlOther = Strings.localized("song_line__bottles_of_beer__pl_other", value: "%1$d bottles of beer on the wall, %1$d bottles of beer.")
    }
    
    public struct WeirdCharacters {
        static let WeirdCharactersExample1 = Strings.localized("weird_characters__example_1", value: "Rock &#39;n&#39; Roll")
        static let WeirdCharactersExample2 = Strings.localized("weird_characters__example_2", value: "Questions &amp; Answers")
        static let WeirdCharactersExample3 = Strings.localized("weird_characters__example_3", value: "What&#39;s \&#34;This\&#34;")
        static let WeirdCharactersExample4 = Strings.localized("weird_characters__example_4", value: "Some %1$@ iOS style string, %@ or %2$@")
        static let WeirdCharactersExample5 = Strings.localized("weird_characters__example_5", value: "Some &lt;a href=\&#34;http://www.google.com\&#34;&gt;Link&lt;/a&gt;")
    }
    

    public static func localized(_ key: String, tableName: String? = nil, bundle: Bundle = Bundle.main, value: String, comment: String = "") -> String {
        return NSLocalizedString(key, tableName: tableName, bundle: bundle, value: value, comment: comment)
    }
}
