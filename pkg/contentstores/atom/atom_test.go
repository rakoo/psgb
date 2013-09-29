package atom

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestAddNewContent(t *testing.T) {
	input, err := ioutil.ReadFile("linuxfr.org-journaux1.atom")
	expectedDate := "2013-09-29T16:08:54+02:00"

	atomStore := NewStore()

	// Test we can add content
	lastid, err := atomStore.AddNewContent("topic1", string(input))
	if err != nil {
		t.Fatal(err)
	}

	if lastid != expectedDate {
		t.Fatalf("Didn't parse the date correctly; expected %s, got %s\n", expectedDate, lastid)
	}

	// After the last one, there should be nothing
	noContent, lastId := atomStore.ContentAfter("topic1", "2013-09-30T16:08:54+02:00")
	expectednocontent := `<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom" xmlns:wfw="http://wellformedweb.org/CommentAPI/">
  <id>tag:linuxfr.org,2005:/journaux</id>
  <link rel="alternate" type="text/html" href="http://linuxfr.org/journaux"/>
  <link rel="self" type="application/atom+xml" href="http://linuxfr.org/journaux.atom"/>
  <title>LinuxFr.org : les journaux</title>
  <updated>2013-09-29T16:08:54+02:00</updated>
  <icon>/favicon.png</icon>
</feed>`
	if noContent != expectednocontent || lastId != "2013-09-29T16:08:54+02:00" {
		t.Fatal("Wrong response where there shouldn't be new entries")
	}

	// After the first one, there should only be one
	oneitem, lastId := atomStore.ContentAfter("topic1", "2013-09-28T21:26:01+02:00")
	expectedoneitem := `<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom" xmlns:wfw="http://wellformedweb.org/CommentAPI/">
  <id>tag:linuxfr.org,2005:/journaux</id>
  <link rel="alternate" type="text/html" href="http://linuxfr.org/journaux"/>
  <link rel="self" type="application/atom+xml" href="http://linuxfr.org/journaux.atom"/>
  <title>LinuxFr.org : les journaux</title>
  <updated>2013-09-29T16:08:54+02:00</updated>
  <icon>/favicon.png</icon>
  <entry>
    <id>tag:linuxfr.org,2005:Diary/34353</id>
    <published>2013-09-29T16:08:54+02:00</published>
    <updated>2013-09-29T16:08:54+02:00</updated>
    <link rel="alternate" type="text/html" href="http://linuxfr.org/users/milien-ri/journaux/c--2"/>
    <title>C(++) ?</title>
    <rights>Licence CC by-sa http://creativecommons.org/licenses/by-sa/3.0/deed.fr</rights>
    <content type="html">&lt;p&gt;Bonjour nal,&lt;/p&gt;

&lt;p&gt;Je t'écris aujourd'hui pour que tu m'aides à trancher une question qui me turlupine. Je suis sur le point de démarrer un projet et je n'arrive pas à choisir entre le C et le C++. Lorsque je développe en C j'ai pour habitude d’appeler mes fonctions xxx_Fonction1, xxx_Fonction2 où xxx est le nom du type (par exemple Socket_Read()).&lt;/p&gt;

&lt;p&gt;Je souhaite aujourd'hui me lancer dans un projet où je devrai gérer de l'héritage. J’aimerais avoir ton avis sur le langage à utiliser :&lt;/p&gt;

&lt;ul&gt;
&lt;li&gt;&lt;p&gt;Le C comme aujourd'hui&lt;/p&gt;&lt;/li&gt;
&lt;li&gt;
&lt;p&gt;Le C avec les pointeurs de fonctions dans la structure&lt;/p&gt;

&lt;pre&gt;&lt;code&gt;struct Socket
{
    (int)(*open)(const char * address, short port);
};
Socket s...;
s.open("www.linuxfr.org",443);
&lt;/code&gt;&lt;/pre&gt;
&lt;/li&gt;
&lt;li&gt;&lt;p&gt;Le C++&lt;/p&gt;&lt;/li&gt;
&lt;/ul&gt;&lt;p&gt;Quel est ton avis cher journal?&lt;/p&gt;&lt;div&gt;&lt;a href="http://linuxfr.org/users/milien-ri/journaux/c--2.epub"&gt;Télécharger ce contenu au format Epub&lt;/a&gt;&lt;/div&gt;    &lt;p&gt;&lt;a href="http://linuxfr.org/users/milien-ri/journaux/c--2#comments"&gt;Lire les commentaires&lt;/a&gt;&lt;/p&gt;
</content>
    <author>
      <name>EmilienR</name>
    </author>
    <category term="c"/>
    <wfw:commentRss>http://linuxfr.org/nodes/99797/comments.atom</wfw:commentRss>
  </entry>
</feed>`
	if oneitem != expectedoneitem || lastId != "2013-09-29T16:08:54+02:00" {
		t.Fatal("Couldn't get only the last item")
	}

	twoitems, lastId := atomStore.ContentAfter("topic1", "2013-09-28T20:26:01+02:00")
	expectedtwoitems := strings.TrimSpace(string(input))
	if twoitems != expectedtwoitems || lastId != "2013-09-29T16:08:54+02:00" {
		t.Log(rune(expectedtwoitems[len(expectedtwoitems)-1]))
		t.Fatal("Couldn't get the 2 items")
	}
}

func TestReAddNewContent(t *testing.T) {
	// Input with entries A and B
	input1, err := ioutil.ReadFile("linuxfr.org-journaux1.atom")

	// Input with entries B and C
	input2, err := ioutil.ReadFile("linuxfr.org-journaux2.atom")
	expectedDate := "2013-09-29T16:40:14+02:00"

	atomStore := NewStore()
	_, err = atomStore.AddNewContent("topic1", string(input1))
	lastid, err := atomStore.AddNewContent("topic1", string(input2))

	if err != nil {
		t.Fatal("Error when parsing content the 2nd time: ", err)
	}

	if lastid != expectedDate {
		t.Fatalf("Didn't parse the date correctly; expected %s, got %s\n", expectedDate, lastid)
	}

	expectedthreeitems := `<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom" xmlns:wfw="http://wellformedweb.org/CommentAPI/">
  <id>tag:linuxfr.org,2005:/journaux</id>
  <link rel="alternate" type="text/html" href="http://linuxfr.org/journaux"/>
  <link rel="self" type="application/atom+xml" href="http://linuxfr.org/journaux.atom"/>
  <title>LinuxFr.org : les journaux</title>
  <updated>2013-09-29T16:40:14+02:00</updated>
  <icon>/favicon.png</icon>
  <entry>
    <id>tag:linuxfr.org,2005:Diary/34354</id>
    <published>2013-09-29T16:40:14+02:00</published>
    <updated>2013-09-29T16:40:14+02:00</updated>
    <link rel="alternate" type="text/html" href="http://linuxfr.org/users/_moumoutte/journaux/sqla_helpers-quelques-trucs-en-vrac"/>
    <title>sqla_helpers : Quelques trucs en vrac</title>
    <rights>Licence CC by-sa http://creativecommons.org/licenses/by-sa/3.0/deed.fr</rights>
    <content type="html">&lt;p&gt;Salut Nal',&lt;/p&gt;

&lt;p&gt;Décidément, si je prend la plume c'est toujours pour t'écrire autour du même sujet. Je me doute que tout cela t'ennuie, mais à moi ça me &lt;s&gt;fait de la pub&lt;/s&gt; du bien et crois moi, je t'en suis très reconnaissant. Alors, je vais te parler de mon tout petit projet sur lequel de temps en temps j'écris quelques lignes de code.&lt;/p&gt;

&lt;p&gt;Pour rappel, il s'agit &lt;a href="http://sqla-helpers.readthedocs.org/en/latest/sqla_helpers.html"&gt;sqla_helpers&lt;/a&gt; qui a pour première fonctionnalité de fournir du sucre syntaxique pour interroger ta base favorite à travers d'objets python et tout cela , grâce à ce merveilleux ORM qu'est SQLAlchemy.&lt;/p&gt;

&lt;p&gt;Allez, je vais tenter de te mettre l'eau à la bouche si tu ne connais pas ou si tu ne t'en souviens pas :&lt;/p&gt;

&lt;p&gt;Avec SQLAlchemy&lt;/p&gt;

&lt;pre&gt;&lt;code&gt;session.query(Treatment).filter(Treatment.name=='bar').one()
session.query(Treatment).join(Status, Status.id == Treatment.status_id).filter(Status.name == u'SUCCESS').all()
# Oui, on pourrait utiliser la jointure implicite ce qui donnerait
session.query(Treatment).filter(Status.name == u'SUCCESS').all()
&lt;/code&gt;&lt;/pre&gt;

&lt;p&gt;Avec sqla_helpers&lt;/p&gt;

&lt;pre&gt;&lt;code&gt;Treatment.get(foo='bar')
Treatment.filter(status__name=u'SUCCESS')
&lt;/code&gt;&lt;/pre&gt;

&lt;p&gt;A l'époque j'avais écrit ça pour le boulot, venant du monde de Django et commençant à travailler avec SQLAlchemy. L'utilité de ce machin là est à mon sens hyper subjectif. On aime ou on aime pas. J'ai continué à le maintenir à peu près et à lui apporter quelques petites bricoles, sachant que un ou deux projets s'en servent.&lt;/p&gt;

&lt;p&gt;En 6 mois d'existence, on a réussi à trouver un petit bug qui a rapidement était corrigé mais d'un point de vue fonctionnalités, y'a pas grand chose à rajouter, sachant que tout le boulot est bien sûr délégué à SQLAlchemy, pas question de faire son taf, il le fait très bien.&lt;/p&gt;

&lt;p&gt;J'en arrive donc à toutes les questions que j'ai pour toi cher journal,&lt;/p&gt;
&lt;h3 id="stabilisation-du-logiciel"&gt;Stabilisation du logiciel&lt;/h3&gt;

&lt;p&gt;Aujourd'hui je viens de publier la version 0.4.0 et en 6 mois d'existence, il y a vraiment peu eu d'évolution de fonctionnalités. Alors je te le demande cher journal, est-ce qu'il faut stabiliser la version et passer en 1.0.0 pour de bon ?&lt;/p&gt;
&lt;h3 id="appel-à-la-traduction"&gt;Appel à la traduction&lt;/h3&gt;

&lt;p&gt;Aujourd'hui, il y a une documentation en anglais. Initialement, tout était en français , car je préférerai avoir une bonne doc en français qu'un truc incompréhensible en anglais. Un collègue à l'époque avait fait cette traduction pour moi, et je l'en remercie, car moi même je n'aurai pas trouver le courage de le faire. Seulement aujourd'hui, cette traduction contient , je pense certain passage encore un peu  flou. Oui, il n'est pas bilingue malheureusement et en plus de ça l'erreur est humaine. Donc, j'ai vue quelques erreurs (dictonnary..) que j'ai corrigé, mais je suis bien trop mauvais pour être exhaustif. Et je pense qu'il y certains passages dont la signification est encore un peu fausse . (Attention, je ne suis pas en train de cracher sur son travail, il a été super sympa de me la faire cette traduction, sans lui tout serait encore en français).&lt;/p&gt;

&lt;p&gt;Donc,  si tu es bon en anglais et est-ce que tu aurais l'amabilité de faire un tour sur la documentation et selon ton bon vouloir, soit me remonter ce qui ne va pas (orthographe, syntaxe, expression, flou…) soit carrément faire un commit sur le projet ( que je serais ravis d'accueillir !)&lt;/p&gt;
&lt;h3 id="le-logiciel-est-il-utilisé"&gt;Le logiciel est-il utilisé ?&lt;/h3&gt;

&lt;p&gt;Je sais que un ou deux projets s'en servent. Je n'avais aucune intention de révolutionner le monde et de faire de ce truc là un standard. Sauf que voilà, quand je regarde sur pypi.python.org, il me dit qu'il y a peu près 560 téléchargements qui ont été fait le mois derniers (sachant qu'il est sortie depuis février 2013). Je sais bien qu'entre les builds automatiques et les upgrades de version, ça ne fait pas 500 utilisateurs tous les mois. Mais à ton avis, cher journal, est-ce que c'est un indicateur auquel on peut se fier ? Sachant que personne n'est venu me demander quoi que ce soit à propos du projet (bug, moyen d'utilisation, doc, etc).&lt;/p&gt;
&lt;h2 id="bookmark"&gt;Bookmark&lt;/h2&gt;

&lt;p&gt;Doc : &lt;a href="http://sqla-helpers.readthedocs.org/en/latest/sqla_helpers.html"&gt;http://sqla-helpers.readthedocs.org/en/latest/sqla_helpers.html&lt;/a&gt;&lt;br&gt;
Github : &lt;a href="https://github.com/moumoutte/sqla_helpers"&gt;https://github.com/moumoutte/sqla_helpers&lt;/a&gt;&lt;/p&gt;&lt;div&gt;&lt;a href="http://linuxfr.org/users/_moumoutte/journaux/sqla_helpers-quelques-trucs-en-vrac.epub"&gt;Télécharger ce contenu au format Epub&lt;/a&gt;&lt;/div&gt;    &lt;p&gt;&lt;a href="http://linuxfr.org/users/_moumoutte/journaux/sqla_helpers-quelques-trucs-en-vrac#comments"&gt;Lire les commentaires&lt;/a&gt;&lt;/p&gt;
</content>
    <author>
      <name>Guillaume Camera</name>
    </author>
    <wfw:commentRss>http://linuxfr.org/nodes/99798/comments.atom</wfw:commentRss>
  </entry>
  <entry>
    <id>tag:linuxfr.org,2005:Diary/34353</id>
    <published>2013-09-29T16:08:54+02:00</published>
    <updated>2013-09-29T16:08:54+02:00</updated>
    <link rel="alternate" type="text/html" href="http://linuxfr.org/users/milien-ri/journaux/c--2"/>
    <title>C(++) ?</title>
    <rights>Licence CC by-sa http://creativecommons.org/licenses/by-sa/3.0/deed.fr</rights>
    <content type="html">&lt;p&gt;Bonjour nal,&lt;/p&gt;

&lt;p&gt;Je t'écris aujourd'hui pour que tu m'aides à trancher une question qui me turlupine. Je suis sur le point de démarrer un projet et je n'arrive pas à choisir entre le C et le C++. Lorsque je développe en C j'ai pour habitude d’appeler mes fonctions xxx_Fonction1, xxx_Fonction2 où xxx est le nom du type (par exemple Socket_Read()).&lt;/p&gt;

&lt;p&gt;Je souhaite aujourd'hui me lancer dans un projet où je devrai gérer de l'héritage. J’aimerais avoir ton avis sur le langage à utiliser :&lt;/p&gt;

&lt;ul&gt;
&lt;li&gt;&lt;p&gt;Le C comme aujourd'hui&lt;/p&gt;&lt;/li&gt;
&lt;li&gt;
&lt;p&gt;Le C avec les pointeurs de fonctions dans la structure&lt;/p&gt;

&lt;pre&gt;&lt;code&gt;struct Socket
{
    (int)(*open)(const char * address, short port);
};
Socket s...;
s.open("www.linuxfr.org",443);
&lt;/code&gt;&lt;/pre&gt;
&lt;/li&gt;
&lt;li&gt;&lt;p&gt;Le C++&lt;/p&gt;&lt;/li&gt;
&lt;/ul&gt;&lt;p&gt;Quel est ton avis cher journal?&lt;/p&gt;&lt;div&gt;&lt;a href="http://linuxfr.org/users/milien-ri/journaux/c--2.epub"&gt;Télécharger ce contenu au format Epub&lt;/a&gt;&lt;/div&gt;    &lt;p&gt;&lt;a href="http://linuxfr.org/users/milien-ri/journaux/c--2#comments"&gt;Lire les commentaires&lt;/a&gt;&lt;/p&gt;
</content>
    <author>
      <name>EmilienR</name>
    </author>
    <category term="c"/>
    <wfw:commentRss>http://linuxfr.org/nodes/99797/comments.atom</wfw:commentRss>
  </entry>
  <entry>
    <id>tag:linuxfr.org,2005:Diary/34352</id>
    <published>2013-09-28T20:26:01+02:00</published>
    <updated>2013-09-28T20:26:01+02:00</updated>
    <link rel="alternate" type="text/html" href="http://linuxfr.org/users/enclair/journaux/sorties-de-gnu-hurd-0-5-gnu-guix-0-4"/>
    <title>Sorties de GNU Hurd 0.5, GNU Guix 0.4</title>
    <rights>Licence CC by-sa http://creativecommons.org/licenses/by-sa/3.0/deed.fr</rights>
    <content type="html">&lt;p&gt;Seize ans après la version 0.2, GNU Hurd, le noyau du projet GNU fait un saut en version 0.5.&lt;br&gt;
Parmi les nouveautés en vrac que j'arrive à peu près à comprendre : support de l'IPV6, utilisation des threads POSIX, "translator" pour lire les CDROMs au format ISO9660, "translator" pour le système de fichier tmpfs, support de /etc/shadow etc…&lt;br&gt;
Il y a également une nouvelle version de GNU Mach, le micro-noyau sur lequel se base actuellement le Hurd : version 1.4, 11 ans après la version 1.3 (dont l'annonce[1] disait que la branche 1.x n'était plus activement développée…)&lt;/p&gt;

&lt;p&gt;GNU Guix, système de paquets et distribution du système GNU, sort en version 0.4 et propose une image pour machine virtuelle QEMU. Guix utilise le noyau Linux (sans ses parties non libres) et le système d'init dmd. Actuellement on ne peut utiliser Guix qu'à partir d'une distribution Linux ou de l'image QEMU, mais d'après l'annonce la prochaine version devrait être amorçable.&lt;/p&gt;

&lt;p&gt;Liens vers les annonces :&lt;br&gt;
GNU Hurd 0.5 &lt;a href="http://lists.gnu.org/archive/html/bug-hurd/2013-09/msg00264.html"&gt;http://lists.gnu.org/archive/html/bug-hurd/2013-09/msg00264.html&lt;/a&gt;&lt;br&gt;
GNU Mach 1.4 &lt;a href="http://lists.gnu.org/archive/html/bug-hurd/2013-09/msg00263.html"&gt;http://lists.gnu.org/archive/html/bug-hurd/2013-09/msg00263.html&lt;/a&gt;&lt;br&gt;
Guix 0.4 &lt;a href="http://lists.gnu.org/archive/html/guix-devel/2013-09/msg00235.html"&gt;http://lists.gnu.org/archive/html/guix-devel/2013-09/msg00235.html&lt;/a&gt;&lt;/p&gt;

&lt;p&gt;[1] &lt;a href="http://git.savannah.gnu.org/cgit/hurd/gnumach.git/tree/%3dannounce-1.3"&gt;http://git.savannah.gnu.org/cgit/hurd/gnumach.git/tree/%3dannounce-1.3&lt;/a&gt;&lt;/p&gt;&lt;div&gt;&lt;a href="http://linuxfr.org/users/enclair/journaux/sorties-de-gnu-hurd-0-5-gnu-guix-0-4.epub"&gt;Télécharger ce contenu au format Epub&lt;/a&gt;&lt;/div&gt;    &lt;p&gt;&lt;a href="http://linuxfr.org/users/enclair/journaux/sorties-de-gnu-hurd-0-5-gnu-guix-0-4#comments"&gt;Lire les commentaires&lt;/a&gt;&lt;/p&gt;
</content>
    <author>
      <name>enclair</name>
    </author>
    <category term="hurd"/>
    <category term="gnu_hurd"/>
    <wfw:commentRss>http://linuxfr.org/nodes/99792/comments.atom</wfw:commentRss>
  </entry>
</feed>`

	actualthreeitems, actualLast := atomStore.ContentAfter("topic1", "")
	if actualthreeitems != expectedthreeitems || actualLast != "2013-09-29T16:40:14+02:00" {
		fmt.Println(actualthreeitems)
		t.Fatal("Couldn't get the 3 items")
	}
}
