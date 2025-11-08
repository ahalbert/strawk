import {EditorView, basicSetup} from "codemirror"
import { EditorState, Compartment } from '@codemirror/state';
import {parser} from "./strawk.parser.js"
import {foldNodeProp, foldInside, indentNodeProp} from "@codemirror/language"
import {styleTags, tags as t} from "@lezer/highlight"
import {LRLanguage} from "@codemirror/language"
import {completeFromList} from "@codemirror/autocomplete"
import {LanguageSupport} from "@codemirror/language"

// Dark theme configuration for CodeMirror to match Bootstrap dark mode
const darkTheme = EditorView.theme({
  "&": {
    backgroundColor: "#212529",
    color: "#adb5bd"
  },
  ".cm-content": {
    caretColor: "#adb5bd"
  },
  ".cm-cursor, .cm-dropCursor": {
    borderLeftColor: "#adb5bd"
  },
  "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection": {
    backgroundColor: "#495057"
  },
  ".cm-activeLine": {
    backgroundColor: "#2c3034"
  },
  ".cm-gutters": {
    backgroundColor: "#1a1d20",
    color: "#6c757d",
    border: "none"
  },
  ".cm-activeLineGutter": {
    backgroundColor: "#2c3034"
  }
}, {dark: true});

// Theme compartment for dynamic switching
const themeCompartment = new Compartment();

export const parserWithMetadata = parser.configure({
  props: [
    styleTags({
      "BEGIN END do while for in continue print next if else break length sub gsub split toupper tolower substr index match" : t.keyword,
      identifier: t.tagName,
      stateidentifier: t.variableName,
      String: t.string,
      Regex: t.regexp,
      Boolean: t.bool,
      String: t.string,
      LineComment: t.lineComment
    }),
    indentNodeProp.add({
      Application: context => context.column(context.node.from) + context.unit
    }),
    foldNodeProp.add({
      Application: foldInside
    })
  ]
});



export const exampleLanguage = LRLanguage.define({
  name: "strawk",
  parser: parserWithMetadata,
  languageData: {
        commentTokens: {line: "#"}
  }
});

export function strawk() {
    return new LanguageSupport(exampleLanguage,  []);
}

function createEditorStateForStrawk(initialContents, options = {}) {
    const isDark = document.documentElement.getAttribute('data-bs-theme') === 'dark';

    let extensions = [
      basicSetup,
      strawk(),
      themeCompartment.of(isDark ? darkTheme : [])
    ];

    return EditorState.create({
        doc: initialContents,
        extensions
    });
}

function createEditorState(initialContents, options = {}) {
    const isDark = document.documentElement.getAttribute('data-bs-theme') === 'dark';

    let extensions = [
      basicSetup,
      themeCompartment.of(isDark ? darkTheme : [])
    ];

    return EditorState.create({
        doc: initialContents,
        extensions
    });
}

function createEditorView(state, parent) {
    return new EditorView({ state, parent });
}

// Function to update the theme of an existing editor
function updateEditorTheme(view, isDark) {
    view.dispatch({
        effects: themeCompartment.reconfigure(isDark ? darkTheme : [])
    });
}

export { createEditorStateForStrawk, createEditorState, createEditorView, updateEditorTheme, themeCompartment };


