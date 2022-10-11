# Idee
Die Stadt- und Stadtteilbibliotheken in Leipzig haben zahlreiche aktuelle Videospiele für Switch, PS5, XBox usw. im Katalog. Leider ist der WebOPAC-Katalog recht sperrig zu benutzen und es ist mühsam herauszufinden, welche Spiele in welcher Bibliothek derzeit ausleihbar sind. Spiele sind idR einer bestimmten Zweigstelle zugeordnet, werden vom OPAC aber auch über andere Zweigstellen als "woanders verfügbar" angezeigt. Darüber hinaus, wechselt das Angebot regelmäßig.

Dieses Programm, soll es ermöglichen, für eine Stadt- bzw. Stadtteilbibliothek alle Videospiele einer bestimmten Plattform anzuzeigen die aktuell verfügbar sind.

# Datenquelle
Die Datengrundlage ist der WebOPAC-Katalog der Leipziger Stadibibliotheken in der "Erweiterten-Suche" unter `https://webopac.stadtbibliothek-leipzig.de/webOPACClient/search.do?methodToCall=switchSearchPage&SearchType=2`

## Session
Die JSP-Serverseite arbeitet mit Session-Ids. Um automatisiert Suchanfragen zu stellen, muss immer zuerst eine gültige Session erzeugt werden.

GET: `https://webopac.stadtbibliothek-leipzig.de/webOPACClient`

Setzt zwei Cookies: 
* USERSESSIONID
* JSESSIONID

## Suche
Das Ziel ist es, alle Einträge im Katalog zu finden, bei denen es sich um ein Videospiel einer bestimmten Plattform zu finden und anschließend die ausleihbaren Ergebnisse zurück zu liefern. Die "Erweiterte Suche" des WebOPAC Katalogs ist für die gezielte Suche einzelner Bücher ausgelegt. Kategorien,wie "Videospiel", gibt es nicht, jedoch sind die Medien mit Schlüsselworten versehen. So existiert im Index der Schlüsselworte die entsprechende Plattform, also:
* Nintendo Switch
* Xbox Series X / One
* Playstation 4/5

Die Konkrete Suchanfrage setzt sich aus einem Basis-Methodenaufruf und den Suchkriterien zusammen. Da die erweiterte Suche die Kombination, mehrerer Parameter ermöglicht, ist die Parameterliste etwas umständlich und lang.
Nachfolgend werden die Suchparameter und die für die Schlüsselwort-Suche relevanten Parameter erläutert.

GET `https://webopac.stadtbibliothek-leipzig.de/webOPACClient/search.do?methodToCall=submit&methodToCallParameter=submitSearch&searchCategories%5B0%5D=902&submitSearch=Suchen&callingPage=searchPreferences`

Weitere Parameter: 
|Parameter                  | Beschreibung                          | Beispiel              |
|-                          |-                                      | -                     |
|CSId                       | USERSESSIONID                         | 1991N87S0583b9ce8380deec85603fd2da7803777dc9d087 |
|searchString               | Schlüsselwort für die Suche           | Nintendo+Switch       |
|selectedViewBranchlib      | Bibliothekszweigstelle für ???        | 0 (Stadtbibliothek)   |
|selectedSearchBranchlib    | Bibliothekszweigstelle für Abholung   | 41 (Gohlis)           |
|timeOut                    | Timeout der Suchanfrage in Sekunden   | 20                    |
|numberOfHits               | Anzahl der Ergebnisse je Seite        | 100                   |

Volständiges Beispiel:

```https://webopac.stadtbibliothek-leipzig.de/webOPACClient/search.do?methodToCall=submit&methodToCallParameter=submitSearch&searchCategories%5B0%5D=902&submitSearch=Suchen&callingPage=searchPreferences&CSId=1991N87S0583b9ce8380deec85603fd2da7803777dc9d087&searchString%5B0%5D=Nintendo+Switch&numberOfHits=500&timeOut=20&selectedViewBranchlib=41&selectedSearchBranchlib=41```

### Kodierung der Stadtteilbibliotheken

|Code   | Bibliothek                    |
|-      |-                              |
|0      |Stadtbibliothek                |
|20     |Bibliothek Plagwitz            |
|21     |Bibliothek Wiederitzsch        |
|22     |Bibliothek Böhlitz-Ehrenberg   |
|23     |Bibl. Lützschena-Stahmeln      |   							
|25     |Bibliothek Holzhausen          |   							
|30     |Bibliothek Südvorstadt         |   							
|41     |Bibliothek Gohlis              |   							
|50     |Bibliothek Volkmarsdorf        |   							
|51     |Bibliothek Schönefeld          |   							
|60     |Bibliothek Paunsdorf           |   							
|61     |Bibliothek Reudnitz            |   							
|70     |Bibliothek Mockau              |   							
|82     |Bibliothek Grünau-Mitte        |   							
|83     |Bibliothek Grünau-Nord         |   							
|84     |Bibliothek Grünau-Süd          |   							
|90     |Fahrbibliothek                 |   		
