## Localization Util

A util to export localized strings from Google Sheets to the xml (Android) or strings (iOS) format.

### Example

An example export of [this sheet][examplesheet] can be seen in the [example](example) folder, or you can try it yourself by running the following command, which will create an `/exports` folder in your current directory.

    [localization] --sheetID 1upHiDHWu5m30tYdhMDP4GXheOWUE4r3VrHfmAUXiuyI android
    [localization] --sheetID 1upHiDHWu5m30tYdhMDP4GXheOWUE4r3VrHfmAUXiuyI ios

The example shows the 

### Usage

Create a new Google Sheet with the following columns: `key, value, android, ios, comment` and fill in your translations.

|key|value|android|ios|comment|
|---|---|---|---|---|
|app_name|Localization Util|||app name|

Value is the default value for your string, but can be overridden per platform in the respective column. While the default value might get escaped or modified during the export, the platform column will be written as-is. Comments are just comments.

Copy the sheet id from the url (`https://docs.google.com/spreadsheets/d/{{sheet_id}}/edit#gid=0`) and run the util to export your strings (default arguments will export to a `/exports` folder)

    [localization] --sheetID {{sheet_id}} [ios|android]

On Android this will result in a `generated_strings.xml` with the following content:

    <?xml version=\"1.0\" encoding=\"utf-8\"?>
    <resources>
        <!-- app name -->
        <string name="app_name">Localization Util</string>
    </resources>
    
While on iOS you get `LocalizableGen.strings`

    /* app name */
    "app_name" = "Localization Util";

Along with a swift helper file `Strings.swift`

    public struct Strings {
        public struct App {
            static let Name = Strings.localized("app_name", value: "Localization Util", comment: "app name")
        }


        public static func localized(_ key: String, tableName: String? = nil, bundle: Bundle = Bundle.main, value: String, comment: String = "") -> String {
            return NSLocalizedString(key, tableName: tableName, bundle: bundle, value: value, comment: comment)
        }
    }


#### Longer Names

Keys like `base_app_name` can be supported for the iOS export by using 2 underscores `__` to signal the end of the group name. `base_app__name` will generate a `struct BaseApp` for iOS.

#### Plurals

Plurals are supported with the `__pl_[<one|other|etc>]` suffix and generate `<plural>` on Android and a `LocalizableGen.stringsdict` on iOS.

#### Localization

Your base language should be in a `default` sheet as you can see in the [example][examplesheet] and any translations should go into sheets with the appropriate name, e.g. `de`. The tool will export the files in the correct folder structure.



  [examplesheet]:https://docs.google.com/spreadsheets/d/1upHiDHWu5m30tYdhMDP4GXheOWUE4r3VrHfmAUXiuyI
