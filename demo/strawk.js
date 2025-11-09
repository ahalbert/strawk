const introduction = `# Welcome to strawk: Structured AWK\n# This variant of the AWK programming language does not iterate over records\n# segmented by newlines. Rather, it matches the input to the regexes below,\n# and runs the actions when the input matches the regex\n\n# Strawk was inspired by Rob Pike's paper "Structural Regular Expressions"\n# You can read it here: https://doc.cat-v.org/bell_labs/structural_regexps/se.pdf\n\n# This program takes the right hand smiley face and converts it\n# to a rectangle coordinates system\n\n#Feedback is always appreciated, please email armand.halbert@gmail.com\n\nBEGIN { \n  x=1\n  y=1 \n}\n\n# If strawk matches this regex (at least one space), it will expand the input\n# to the maximum matching string and then run the rule.\n/ +/ {\n  x += length($0)\n}\n\n\n/#+/ {\n  print "rect",x, x+length($0),y, y+1\n  x+=length($0)\n}\n\n/\\n/ {\n  x=1\n  y++ \n}\n`
const introductioninput = "                          ####################\n                       ###########################\n                    #################################         ##   #####\n    ######        ######################################       ##########\n ## ######      ##########    #############    ##########       ########\n #########     ##########      ###########      ###########    ########\n   #######    ###########      ###########      #######################\n   #######################    #############    ##############  ######\n    #########################################################     ####\n     ###   ###################################################     #####\n    ####   ###################################################       ####\n    ###    ############################################## #################\n   #############  #####################################   ##################\n   #############   ##################################     ############\n  ####       ####    ##############################      ####\n             #####     #########################         ###\n               ####          ###############           ####\n                #####                                #####\n                 ######      ##############        #####\n                   ########     #############   #######\n                      ###########  #################\n                         ######################\n                                 ###############\n                                     ############\n                                      ###########\n                                       ########";

const sentences ='BEGIN {\n  count=0\n  words=0\n}\n# This regex matches each sentence in the input, and then prints the sentence. It works across newlines!\n/(?s)(.*)\\./ { \n  sentence = gsub(/\\n/, "")  #gsub (Global SUBstitution) replaces all newlines in the string with an empty string\n  sentence = sub(/^ /, "", sentence) #sub only replaces the first occurance of a regex\n  print sentence\n  splitwords = split(sentence, " ") # split splits a string into an array based on an argument\n  count += 1\n  words += length(splitwords)\n}\n\nEND { print "Average length of sentence in text:",words / count }\n';
const sentencesInput = "Sunt molestias autem doloremque. Sed ut doloremque occaecati quo est quam numquam exercitationem suscipit et. \nAd enim voluptatem consequatur vitae quis. Maxime quasi magni velit eius nam aut esse voluptatibus quis velit \nrepellendus. Temporibus facilis ut porro deleniti excepturi quas alias placeat. Numquam minus aut doloribus \nfugit magni dolorum. Omnis ut minus rem quo est qui voluptate iste impedit. Vel impedit qui qui sit explicabo \nassumenda recusandae voluptatem quia animi.\n";

const paragraphs ='BEGIN {\n  paragraph=0\n}\n# This regex matches each paragraph in the input, and then prints every other paragraph.\n/(?s)(.*)\\n\\n/ {\n  paragraph += 1\n  if paragraph % 2 == 0 {\n    print $0\n  }\n}\n';
const paragraphsInput ='Sunt molestias autem doloremque. Sed ut doloremque occaecati quo est quam numquam \nexercitationem suscipit et. Ad enim voluptatem consequatur vitae quis. Maxime \nquasi magni velit eius nam aut esse voluptatibus quis velit repellendus. \nTemporibus facilis ut porro deleniti excepturi quas alias placeat. Numquam minus \naut doloribus fugit magni dolorum. Omnis ut minus rem quo est qui voluptate iste \nimpedit.\n\nId eligendi quis ab aliquid impedit tempore velit corrupti. Voluptates fugiat quia \nalias doloribus voluptatem laboriosam possimus quidem repellendus quo aperiam est \nreiciendis culpa. Occaecati ratione est voluptas accusantium eos qui sint consequatur \nmaxime. Ad voluptatem sed nihil rem est ipsa aut impedit numquam vel voluptas. Officia \nquod sit ullam dolor placeat aliquid harum modi in sunt qui. Ex est id soluta. Modi \ndistinctio ea quia voluptatum corrupti occaecati maxime quia unde voluptatem explicabo \nquam. Porro voluptatem dolor recusandae possimus repudiandae incidunt.\n\nEt nulla rerum deleniti sed labore eos. Distinctio omnis vel illum eos in. Perferendis \nest aliquam saepe dolores fugiat tempore minima molestiae eos adipisci et distinctio \niste.\n\nDignissimos provident voluptatem vero eum blanditiis voluptatum. A quaerat voluptas est \narchitecto modi. Quasi numquam provident consectetur qui deserunt assumenda sequi impedit \nvitae eaque incidunt et. Qui nobis impedit molestiae omnis ducimus et voluptatem quia.\n\nOccaecati sunt dolore incidunt quas eos suscipit nisi quas similique eaque dolorum. Quos \nconsequatur temporibus earum aut dolor aut aut itaque quibusdam quibusdam reiciendis est \ncum. Consequatur at dolorum consequatur nulla at dignissimos ipsam perspiciatis rerum \nnulla enim adipisci veniam ut mollitia. Ducimus non dolorem in fuga quo quo cumque \ncorporis aut ex. Rerum pariatur sunt sint quia perspiciatis et sed illo quam modi numquam \nodio aliquid repellendus molestiae.\n';

const mockingSpongebob ='BEGIN { \n   sentence = 0\n}\n\n# Matches each sentence\n/(?s)(.*)\\./ { \n  sentence++\n  s = sub(/\\n/, "", $1) #Remove newlines\n  if sentence == 4 {\n    new_sentence = ""\n    for (char = 0; char < length(s); char++) {\n      if char % 2  == 0 {\n        c = toupper(substr(s, char, 1))\n      } else {\n        c = tolower(substr(s, char, 1))\n      }\n      new_sentence = new_sentence c # Concatenates new_sentence and c\n    }\n    print new_sentence\n  }\n}\n';
const mockingSpongebobInput = "Sunt molestias autem doloremque. Sed ut doloremque occaecati quo est quam numquam exercitationem suscipit et. \nAd enim voluptatem consequatur vitae quis. Maxime quasi magni velit eius nam aut esse voluptatibus quis velit \nrepellendus. Temporibus facilis ut porro deleniti excepturi quas alias placeat. Numquam minus aut doloribus \nfugit magni dolorum. Omnis ut minus rem quo est qui voluptate iste impedit. Vel impedit qui qui sit explicabo \nassumenda recusandae voluptatem quia animi.\n";

const captureGroups ='# If a regex is matched, the full value is put in $0 and each capture group in $1, $2...\n/Name: (.*)\\nFavorite Food: (.*)\\n?/ {\n   print $1, $2\n   if ($2 in foods) {\n     foods[$2] += 1\n   } else {\n     foods[$2] = 1\n   }\n}\n\nEND {\n   count, mostpopular = 0, ""\n   for (food in foods) {\n     if foods[food] > count {\n       count = foods[food]\n       mostpopular = food\n     }\n   }\n   print "The most popular food is:", mostpopular\n}\n';
const captureGroupsInput ='Name: Bob\nFavorite Food: Apple\n\nName: Alice\nFavorite Food: Banana\n\nName: Pam\nFavorite Food: Orange\n\nName: George\nFavorite Food: Banana';

const fsastate = cm6.createEditorStateForStrawk(introduction);
const fsaed = cm6.createEditorView(fsastate, document.getElementById("strawkeditor"));
const inputstate = cm6.createEditorState(introductioninput);
const inputed = cm6.createEditorView(inputstate, document.getElementById("inputeditor"));
const outputstate = cm6.createEditorState("");
const outputed = cm6.createEditorView(outputstate, document.getElementById("outputeditor"));

function runProgram() {
  var body = {
    "program" : fsaed.state.doc.toString(),
    "data" : inputed.state.doc.toString()
  };
  $.ajax("https://us-east1-ahalbert-clickstream.cloudfunctions.net/strawk-api", {
      method : 'POST',
      data : JSON.stringify(body),
      contentType : 'application/json',
      success: function(data) {
        var newState = cm6.createEditorState(data['output'])
        outputed.setState(newState);
      }
    }
  );
}

function changeExample(fsa, inp) {
  var newState = cm6.createEditorStateForStrawk(fsa);
  fsaed.setState(newState);
  newState = cm6.createEditorState(inp);
  inputed.setState(newState);
}
