#include "TFile.h"
#include "TTree.h"
#include <TTreeReader.h>
#include <TTreeReaderValue.h>
#include <TTreeReaderArray.h>

// Get current date/time, format is YYYY-MM-DD.HH:mm:ss
const std::string currentDateTime() {
    time_t     now = time(0);
    struct tm  tstruct;
    char       buf[80];
    tstruct = *localtime(&now);
    // Visit http://en.cppreference.com/w/cpp/chrono/c/strftime
    // for more information about date/time format
    strftime(buf, sizeof(buf), "%Y-%m-%d %X", &tstruct);

    return buf;
}

TF1* Fit(TH1F* h) {
	Double_t low = 1400;
	Double_t high = 3000;
	TF1* f = 0;
	h->GetXaxis()->SetRange(h->GetXaxis()->FindBin(0.),h->GetXaxis()->FindBin(3000.));
	int maxbin = h->GetMaximumBin();
	double AbscissaAtMax = h->GetXaxis()->GetBinCenter(maxbin);
	//cout << "AbscissaAtMax = " << AbscissaAtMax << endl;
	h->Fit("gaus", "Q", "", AbscissaAtMax - 0.20*AbscissaAtMax, AbscissaAtMax + 0.20*AbscissaAtMax);
	f = h->GetFunction("gaus");
	/*
	for(int i = 0; i < 3; i++) {
		h->Fit("gaus", "Q", "", low, high);
		f = h->GetFunction("gaus");
		//cout << "here1" << endl;
		if(f) {
			Double_t mean = f->GetParameter(1);
			Double_t sigma = f->GetParameter(2);
			//cout << "here2" << endl;
			if(mean == 0) {
				cout << "ERROR: fitted mean is zero" << endl;
				return 0;
			}
			if(sigma > 0.20*2000) {
				sigma = 0.15*2000;
			}
			low = mean - 2 * sigma;
			high = mean + 2 * sigma;
		}
	}
	*/
	return f;
}

void DistribAmplCharge(TString fileName0, TString fileName1="", TString fileName2="", TString fileName3="", TString fileName4="") {
	gStyle->SetOptStat(0);
	gStyle->SetOptTitle(0);
	
	TChain ch("tree");
	ch.Add(fileName0);
	if(fileName1 != "") {
		ch.Add(fileName1);
	}
	if(fileName2 != "") {
		ch.Add(fileName2);
	}
	if(fileName3 != "") {
		ch.Add(fileName3);
	}
	if(fileName4 != "") {
		ch.Add(fileName4);
	}
	
        TTreeReader reader(&ch);
        TTreeReaderValue<UInt_t> Run(reader, "Run");
        TTreeReaderValue<UInt_t> Evt(reader, "Evt");
        TTreeReaderArray<UShort_t> IChanAbs240(reader, "IChanAbs240");
        TTreeReaderArray<Double_t> Ampl(reader, "Ampl");
        TTreeReaderArray<Double_t> Charge(reader, "Charge");

	int Nbins = 200;
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
	
	TCanvas* cLeft = new TCanvas("cLeft", "cLeft", 1500, 800);
	TCanvas* cRight = new TCanvas("cRight", "cRight", 1500, 800);
	cLeft->SetFillColor(7);
	cRight->SetFillColor(kYellow);
	cLeft->Divide(5, 6);
	cRight->Divide(5, 6);
	
	ofstream of("energy.csv");
	of << "# LAPD energy calibration constants (creation date: " << currentDateTime() 
	   << ", input files:" 
	   << " " << fileName0.Data() 
	   << " " << fileName1.Data() 
	   << " " << fileName2.Data()
	   << " " << fileName3.Data()
	   << ")" << endl;
	of << "# Calibration constant defined as the number of ADC counts corresponding to 511 keV" << endl;
        of << "# iChannelAbs240 calibConstant calibConstantError " << endl;
	
	// Draw right hemisphere
	for (int iQ = 0; iQ < 60; iQ++) {
		if(iQ < 30) {
			int irow = iQ/5;
			int icol = 4-iQ%5+1;
			int ipad = irow*5 + icol;
			cRight->cd(ipad);
		} else {
			int iQprime = iQ - 30;
			int irow = 5-iQprime/5;
			int icol = iQprime%5+1;
			int ipad = irow*5 + icol;
			cLeft->cd(ipad);
		}
		gPad->SetFillColor(kWhite);
// 		gPad->SetBottomMargin(0);
// 		gPad->SetTopMargin(0);
// 		gPad->SetLeftMargin(0);
// 		gPad->SetRightMargin(0);
		TLegend* leg = new TLegend(0.6,0.6,1,1);
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
			histos[iChanAbs240]->Scale(1/histos[iChanAbs240]->Integral());
			histos[iChanAbs240]->SetLineColor(color);
			histos[iChanAbs240]->SetLineWidth(1);
			histos[iChanAbs240]->SetFillStyle(3001);
			histos[iChanAbs240]->SetFillColor(color);
			histos[iChanAbs240]->GetXaxis()->SetLabelSize(0.055);
			histos[iChanAbs240]->GetYaxis()->SetLabelSize(0.055);
			if(iC == 0) {
				histos[iChanAbs240]->Draw();
			} else {
				histos[iChanAbs240]->Draw("same");
			}
			cout << "iChanAbs240: " << iChanAbs240 ;
			TF1* fitfunc = 0;
			if(histos[iChanAbs240]->GetEntries() > 0) {
				fitfunc = Fit(histos[iChanAbs240]);
				if(fitfunc) {
					Double_t mean = fitfunc->GetParameter(1);
					Double_t sigma = fitfunc->GetParameter(2);
					Double_t meanErr = fitfunc->GetParError(1);
					cout << " ->  " << mean << "  " << meanErr << "  " << sigma;
					if(sigma < 500 && meanErr < 0.20*mean) {
						of << iChanAbs240 << " " << mean << " " << meanErr << endl;
					} else {
						cout << " <=== Warning !!" ;
					}
					cout << endl;
				}
			} else {
				of << iChanAbs240 << " 0 0" << endl;
			}
			leg->AddEntry(histos[iChanAbs240], Form("iChanAbs240=%i", iChanAbs240), "f");
		}
		leg->SetLineWidth(0);
		leg->Draw();
	}
	of.close();
}
