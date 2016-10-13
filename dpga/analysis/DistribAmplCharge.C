#include "TFile.h"
#include "TTree.h"
#include <TTreeReader.h>
#include <TTreeReaderValue.h>
#include <TTreeReaderArray.h>

void DistribAmplCharge(TString fileName) {
	TFile* f = new TFile(fileName, "read");
	
        TTreeReader reader("tree", f);
        TTreeReaderValue<UInt_t> Run(reader, "Run");
        TTreeReaderValue<UInt_t> Evt(reader, "Evt");
        TTreeReaderArray<UShort_t> IChanAbs240(reader, "IChanAbs240");
        TTreeReaderArray<Double_t> Ampl(reader, "Ampl");
        TTreeReaderArray<Double_t> Charge(reader, "Charge");

	int Nbins = 150;
	float minX = 0;
	float maxX = 4095;
	
	std::vector<TH1F*> histos;
	for(int i = 0; i < 240; ++i) {
		histos.push_back(new TH1F(Form("histo_%i", i), Form("histo_%i", i), Nbins, minX, maxX));
	}
	
        while (reader.Next()) {
                //cout << *Run << " " << *Evt << " " << IChanAbs240[0] <<  " " << IChanAbs240[1] << " " << Ampl[0] << " " << Ampl[1] << endl;
		if(IChanAbs240[0] >= 120) {
			cout << "ERROR: IChanAbs240[0] >= 120" << endl;
			return;
		}
		if(IChanAbs240[1] < 120) {
			cout << "ERROR: IChanAbs240[1] < 120" << endl;
			return;
		}
		histos[IChanAbs240[0]]->Fill(Ampl[0]);
		histos[IChanAbs240[1]]->Fill(Ampl[1]);
        }
	
	TCanvas* cLeft = new TCanvas("cLeft", "cLeft", 800, 800);
	TCanvas* cRight = new TCanvas("cRight", "cRight", 800, 800);
	cLeft->Divide(6, 5);
	cRight->Divide(6, 5);
	
	// Draw right hemisphere
	for (int iQ = 0; iQ < 60; iQ++) {
		if(iQ < 30) {
			cRight->cd(iQ+1);
		} else {
			cLeft->cd(iQ-30+1);
		}
		for (int iC = 0; iC < 4; iC++) {
			int iChanAbs240 = 4*iQ + iC;
			int color;
			if(iC == 0)
				color = kRed;
			else if(iC == 1)
				color = kGreen+2;
			else if(iC == 2)
				color = kBlue;
			else if(iC == 3)
				color = kMagenta;
			histos[iChanAbs240]->SetLineColor(color);
			histos[iChanAbs240]->SetLineWidth(2);
			histos[iChanAbs240]->SetFillStyle(3001);
			histos[iChanAbs240]->SetFillColor(color);
			if(iC == 0) {
				histos[iChanAbs240]->Draw();
			} else {
				histos[iChanAbs240]->Draw("same");
			}
		}
	}
}
