// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//
const ts = require('typescriptServices.js');

var compilerOptions = {
    module: ts.ModuleKind.CommonJS,
};

function transpileModule(input) {
    var ret = ts.transpileModule(input, {
        compilerOptions: compilerOptions,
        reportDiagnostics: true,
        fileName: "_contract.ts",
    });

    if (ret.diagnostics.length > 0) {
        ret.diagnostics.forEach(diagnostic => {
            var message = ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n');

            if (diagnostic.file) {
                var {
                    line,
                    character
                } = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start);
                message = diagnostic.file.fileName + ":" + (line + 1) + ":" + (character + 1) + ": " + message;
            }
            throw new Error("fail to transpile TypeScript: " + message);
        });
    }

    return {
        jsSource: ret.outputText,
        lineOffset: 0,
    };
};

exports.transpileModule = transpileModule;
